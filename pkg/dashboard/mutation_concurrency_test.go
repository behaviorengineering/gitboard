package dashboard

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

type concurrentLocalFake struct {
	common string

	mu           sync.Mutex
	active       int
	maxActive    int
	pullStarted  chan struct{}
	pullRelease  chan struct{}
	fetchStarted chan struct{}
	fetchRelease chan struct{}
	pullCalls    atomic.Int32
	fetchCalls   atomic.Int32
	syncCalls    atomic.Int32
	pruneCalls   atomic.Int32

	indexErr error
	pullErr  error
}

func newConcurrentLocalFake(common string) *concurrentLocalFake {
	return &concurrentLocalFake{
		common:       common,
		pullStarted:  make(chan struct{}),
		pullRelease:  make(chan struct{}),
		fetchStarted: make(chan struct{}),
		fetchRelease: make(chan struct{}),
	}
}

func (f *concurrentLocalFake) enter() {
	f.mu.Lock()
	f.active++
	if f.active > f.maxActive {
		f.maxActive = f.active
	}
	f.mu.Unlock()
}

func (f *concurrentLocalFake) leave() {
	f.mu.Lock()
	f.active--
	f.mu.Unlock()
}

func (f *concurrentLocalFake) MaxActive() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.maxActive
}

func (f *concurrentLocalFake) ScanRoots(context.Context, []string) localgit.Discovery {
	return localgit.Discovery{}
}

func (f *concurrentLocalFake) InspectPath(_ context.Context, path string) localgit.Status {
	return localgit.Status{Mapped: true, Path: path, Branch: "main"}
}

func (f *concurrentLocalFake) OriginSync(context.Context, string) (localgit.RepoSync, error) {
	return localgit.RepoSync{}, nil
}

func (f *concurrentLocalFake) CommonGitDir(context.Context, string) (string, error) {
	return f.common, nil
}

func (f *concurrentLocalFake) ListLocalHeads(context.Context, string) ([]string, error) {
	return nil, nil
}

func (f *concurrentLocalFake) FetchOriginSmart(ctx context.Context, _ string, _ localgit.FetchPolicy, _ *localgit.OriginFetchCache) (bool, error) {
	f.fetchCalls.Add(1)
	f.enter()
	defer f.leave()
	select {
	case <-f.fetchStarted:
	default:
		close(f.fetchStarted)
	}
	select {
	case <-f.fetchRelease:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (f *concurrentLocalFake) OriginRemote(context.Context, string) (string, error) {
	return "https://example.test/acme/app.git", nil
}

func (f *concurrentLocalFake) PullFFOnly(ctx context.Context, _, _ string) (localgit.PullPhases, error) {
	f.pullCalls.Add(1)
	f.enter()
	defer f.leave()
	select {
	case <-f.pullStarted:
	default:
		close(f.pullStarted)
	}
	if f.pullErr != nil {
		return localgit.PullPhases{}, f.pullErr
	}
	select {
	case <-f.pullRelease:
		return localgit.PullPhases{}, nil
	case <-ctx.Done():
		return localgit.PullPhases{}, ctx.Err()
	}
}

func (f *concurrentLocalFake) RemoveSafeCheckout(ctx context.Context, _, _, _ string) error {
	f.pruneCalls.Add(1)
	f.enter()
	defer f.leave()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (f *concurrentLocalFake) EnsureWritableIndex(context.Context, string, bool) error {
	return f.indexErr
}

func (f *concurrentLocalFake) ContentOnDefault(context.Context, string, string, string) (bool, string, error) {
	return false, "", nil
}

func (f *concurrentLocalFake) CompareSync(ctx context.Context, _, _ string) (localgit.SyncInspection, error) {
	f.syncCalls.Add(1)
	select {
	case <-ctx.Done():
		return localgit.SyncInspection{}, ctx.Err()
	default:
		return localgit.SyncInspection{Relation: "behind_only"}, nil
	}
}

func (f *concurrentLocalFake) CanonicalCheckoutPath(_ context.Context, path string) string {
	return path
}

func (f *concurrentLocalFake) Mutator(*localgit.MutationLease) (LocalMutator, error) {
	return f, nil
}

func TestConcurrentPullsSameCommonDirSerialize(t *testing.T) {
	common := filepath.Clean(t.TempDir())
	fake := newConcurrentLocalFake(common)
	close(fake.pullRelease)
	service := newTestService(&fakeExec{}, fake)
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: common,
		}},
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := commands.PullFF(context.Background(), doc, PullFFRequest{
				ProjectID: "gh-app",
				Branch:    "main",
				RepoPath:  common,
			})
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	if fake.MaxActive() != 1 {
		t.Fatalf("max active local mutations = %d, want 1", fake.MaxActive())
	}
	if fake.pullCalls.Load() != 4 {
		t.Fatalf("pull calls = %d", fake.pullCalls.Load())
	}
}

func TestPullVersusOriginFetchSerialize(t *testing.T) {
	common := filepath.Clean(t.TempDir())
	fake := newConcurrentLocalFake(common)
	service := newTestService(&fakeExec{}, fake)
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: common,
		}},
	}

	pullErr := make(chan error, 1)
	go func() {
		_, err := commands.PullFF(context.Background(), doc, PullFFRequest{
			ProjectID: "gh-app",
			Branch:    "main",
			RepoPath:  common,
		})
		pullErr <- err
	}()
	<-fake.pullStarted

	fetchDone := make(chan error, 1)
	go func() {
		fetchDone <- service.fetchOriginLocked(context.Background(), common, common, localgit.FetchPolicy{TTL: time.Minute, Fresh: true, Branches: []string{"main"}})
	}()

	select {
	case <-fake.fetchStarted:
		t.Fatal("fetch must not start while pull holds the coordinator")
	case <-time.After(80 * time.Millisecond):
	}

	close(fake.pullRelease)
	if err := <-pullErr; err != nil {
		t.Fatal(err)
	}
	close(fake.fetchRelease)
	if err := <-fetchDone; err != nil {
		t.Fatal(err)
	}
	if fake.MaxActive() != 1 {
		t.Fatalf("max active = %d, want 1", fake.MaxActive())
	}
}

func TestPullVersusSyncInvestigationSerialize(t *testing.T) {
	common := filepath.Clean(t.TempDir())
	fake := newConcurrentLocalFake(common)
	close(fake.fetchRelease) // investigate fetch completes once the pull lease is released.
	service := newTestService(&fakeExec{}, fake)
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: common,
		}},
	}

	pullErr := make(chan error, 1)
	go func() {
		_, err := commands.PullFF(context.Background(), doc, PullFFRequest{
			ProjectID: "gh-app",
			Branch:    "main",
			RepoPath:  common,
		})
		pullErr <- err
	}()
	<-fake.pullStarted

	syncStarted := make(chan struct{})
	syncErr := make(chan error, 1)
	go func() {
		close(syncStarted)
		_, err := commands.InvestigateSync(context.Background(), doc, SyncInvestigationRequest{
			ProjectID: "gh-app",
			Branch:    "main",
			RepoPath:  common,
		})
		syncErr <- err
	}()
	<-syncStarted
	time.Sleep(50 * time.Millisecond)
	if fake.fetchCalls.Load() != 0 {
		t.Fatal("investigate fetch must wait for pull lease")
	}
	if fake.syncCalls.Load() != 0 {
		t.Fatal("compare must not start before investigate fetch")
	}

	close(fake.pullRelease)
	if err := <-pullErr; err != nil {
		t.Fatal(err)
	}
	if err := <-syncErr; err != nil {
		t.Fatal(err)
	}
	if fake.MaxActive() != 1 {
		t.Fatalf("max active = %d, want 1", fake.MaxActive())
	}
	if fake.fetchCalls.Load() < 1 {
		t.Fatal("investigate should fetch under the lease")
	}
	if fake.syncCalls.Load() < 1 {
		t.Fatal("investigate should compare after the lease")
	}
}

func TestQueuedPullCancelsWithContext(t *testing.T) {
	common := filepath.Clean(t.TempDir())
	fake := newConcurrentLocalFake(common)
	service := newTestService(&fakeExec{}, fake)
	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: common,
		}},
	}

	firstErr := make(chan error, 1)
	go func() {
		_, err := commands.PullFF(context.Background(), doc, PullFFRequest{
			ProjectID: "gh-app",
			Branch:    "main",
			RepoPath:  common,
		})
		firstErr <- err
	}()
	<-fake.pullStarted

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err := commands.PullFF(ctx, doc, PullFFRequest{
		ProjectID: "gh-app",
		Branch:    "main",
		RepoPath:  common,
	})
	if !errors.Is(err, localgit.ErrMutationCanceled) {
		t.Fatalf("want mutation canceled, got %v", err)
	}

	close(fake.pullRelease)
	if err := <-firstErr; err != nil {
		t.Fatal(err)
	}
}

func TestMutationEpochInvalidatesStaleSnapshot(t *testing.T) {
	common := filepath.Clean(t.TempDir())
	fake := newConcurrentLocalFake(common)
	close(fake.pullRelease)
	close(fake.fetchRelease)
	service := newTestService(&fakeExec{}, fake)

	target := inspectTarget{checkout: localgit.Checkout{Path: common, CommonGitDir: common}, common: common}
	before := service.mutationEpochs([]inspectTarget{target})
	if service.mutationEpochChanged(target, before) {
		t.Fatal("epoch should be stable before mutation")
	}

	commands := NewCommands(service)
	doc := boardconfig.File{
		Projects: []boardconfig.Project{{
			ID:        "gh-app",
			Host:      boardconfig.HostGitHub,
			Path:      "acme/app",
			LocalPath: common,
		}},
	}
	if _, err := commands.PullFF(context.Background(), doc, PullFFRequest{
		ProjectID: "gh-app",
		Branch:    "main",
		RepoPath:  common,
	}); err != nil {
		t.Fatal(err)
	}
	if !service.mutationEpochChanged(target, before) {
		t.Fatal("expected epoch change after pull")
	}
}
