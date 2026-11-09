package localgit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/behaviorengineering/gitboard/pkg/cliexec"
)

// fakeExec returns canned git output matched by argv substrings (longest match wins).
type fakeExec struct {
	mu        sync.Mutex
	responses map[string][]byte
	errors    map[string]error
	calls     []string
}

func (f *fakeExec) LookPath(name string) (string, error) {
	return "/fake/" + name, nil
}

func (f *fakeExec) match(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, key)
	var best string
	for substr := range f.responses {
		if strings.Contains(key, substr) && len(substr) >= len(best) {
			best = substr
		}
	}
	for substr := range f.errors {
		if strings.Contains(key, substr) && len(substr) >= len(best) {
			best = substr
		}
	}
	if best == "" {
		return nil, fmt.Errorf("unexpected argv: %s", key)
	}
	if err, ok := f.errors[best]; ok {
		return nil, err
	}
	return f.responses[best], nil
}

func (f *fakeExec) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	return f.match(name, args...)
}

func (f *fakeExec) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	return f.Run(ctx, name, args...)
}

// leasedMutator acquires an in-process lease on common and returns its mutator.
func leasedMutator(t *testing.T, run cliexec.Exec, common string) *Mutator {
	t.Helper()
	lease, err := NewMutationCoordinator(nil).Acquire(context.Background(), common)
	if err != nil {
		t.Fatalf("acquire lease: %v", err)
	}
	t.Cleanup(func() { releaseOrFail(t, lease) })
	m, err := lease.Mutator(run)
	if err != nil {
		t.Fatalf("lease mutator: %v", err)
	}
	return m
}
