package dashboard

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

func TestShellConfigOnly(t *testing.T) {
	doc := config.File{
		Projects: []config.Project{
			{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
			{ID: "gl-lib", Label: "Lib", Host: config.HostGitLab, Path: "acme/lib"},
			{ID: "az-x", Label: "X", Host: config.HostAzureDevOps, Path: "org/proj/x"},
		},
		Views: []config.View{
			{ID: "work", Label: "Work", Projects: []string{"gh-app", "gl-lib"}},
		},
	}
	out, err := Shell(doc, "work")
	if err != nil {
		t.Fatal(err)
	}
	if out.ActiveView != "work" {
		t.Fatalf("active_view: %q", out.ActiveView)
	}
	if len(out.Projects) != 2 {
		t.Fatalf("projects: %d", len(out.Projects))
	}
	byID := map[string]board.ProjectSummary{}
	for _, p := range out.Projects {
		byID[p.ID] = p
		if len(p.Branches) != 0 {
			t.Fatalf("%s branches should be empty: %+v", p.ID, p.Branches)
		}
		if p.Local != nil {
			t.Fatalf("%s local should be nil", p.ID)
		}
		if p.CI != nil {
			t.Fatalf("%s ci should be nil", p.ID)
		}
		if p.OpenURL == "" {
			t.Fatalf("%s missing open_url", p.ID)
		}
	}
	if !byID["gh-app"].Capabilities.FailedJobs {
		t.Fatalf("github capabilities: %+v", byID["gh-app"].Capabilities)
	}
	if !byID["gl-lib"].Capabilities.FailedJobs {
		t.Fatalf("gitlab capabilities: %+v", byID["gl-lib"].Capabilities)
	}
	if _, err := Shell(doc, "nope"); err == nil {
		t.Fatal("expected unknown view error")
	}
}

func TestShellNoForgeExec(t *testing.T) {
	fx := &countingExec{responses: dualHostFake().responses}
	s := New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), nil)
	doc := config.File{
		Projects: []config.Project{
			{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
		},
	}
	_, err := Shell(doc, "")
	if err != nil {
		t.Fatal(err)
	}
	if fx.calls != 0 {
		t.Fatalf("Shell must not call forge exec; calls=%d", fx.calls)
	}
	_, err = s.Collect(context.Background(), doc, true, "")
	if err != nil {
		t.Fatal(err)
	}
	if fx.calls == 0 {
		t.Fatal("expected Collect to call forge exec")
	}
}

type countingExec struct {
	responses map[string][]byte
	mu        sync.Mutex
	calls     int
}

func (f *countingExec) LookPath(name string) (string, error) {
	return "/fake/" + name, nil
}

func (f *countingExec) match(name string, args ...string) ([]byte, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	key := name + " " + strings.Join(args, " ")
	var best string
	for substr := range f.responses {
		if strings.Contains(key, substr) && len(substr) >= len(best) {
			best = substr
		}
	}
	if best == "" {
		return nil, fmt.Errorf("unexpected argv: %s", key)
	}
	return f.responses[best], nil
}

func (f *countingExec) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	return f.match(name, args...)
}

func (f *countingExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := f.Run(ctx, name, args...)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(out)) == "" {
		return []byte("[]"), nil
	}
	return out, nil
}

func TestCollectStreamEmitsBeforeSlowestFinishes(t *testing.T) {
	fx := &delayExec{
		fakeExec: dualHostFake(),
		delayFor: map[string]time.Duration{
			"acme/app": 150 * time.Millisecond,
			"acme/lib": 0,
		},
	}
	s := New(remotegit.NewGitHub(fx), remotegit.NewGitLab(fx), remotegit.NewAzureDevOps(fx), remotegit.NewBitbucket(fx), nil)
	doc := config.File{
		Projects: []config.Project{
			{ID: "gh-app", Label: "App", Host: config.HostGitHub, Path: "acme/app"},
			{ID: "gl-lib", Label: "Lib", Host: config.HostGitLab, Path: "acme/lib"},
		},
	}
	fastSeen := make(chan string, 2)
	var order []string
	var mu sync.Mutex
	errCh := make(chan error, 1)
	go func() {
		_, err := s.CollectStream(context.Background(), doc, CollectOpts{Fresh: true}, func(row board.ProjectSummary) {
			mu.Lock()
			order = append(order, row.ID)
			mu.Unlock()
			fastSeen <- row.ID
		})
		errCh <- err
	}()

	select {
	case <-fastSeen:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first project emit")
	}

	select {
	case <-fastSeen:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for second project emit")
	}

	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 {
		t.Fatalf("emit order: %v", order)
	}
}

type delayExec struct {
	*fakeExec
	delayFor map[string]time.Duration
}

func (d *delayExec) sleepFor(args []string) {
	joined := strings.Join(args, " ")
	for needle, delay := range d.delayFor {
		if strings.Contains(joined, needle) && delay > 0 {
			time.Sleep(delay)
			return
		}
	}
}

func (d *delayExec) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	d.sleepFor(args)
	return d.fakeExec.Run(ctx, name, args...)
}

func (d *delayExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	d.sleepFor(args)
	return d.fakeExec.RunJSON(ctx, name, args...)
}
