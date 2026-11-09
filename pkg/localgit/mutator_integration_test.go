//go:build integration

package localgit_test

import (
	"context"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// testCtx is a per-test context with a deadline: cliexec refuses to run git
// without one.
func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	t.Cleanup(cancel)
	return ctx
}

// mutatorFor leases repoPath's common git directory and returns its mutator.
func mutatorFor(t *testing.T, run cliexec.Exec, repoPath string) *localgit.Mutator {
	t.Helper()
	in := localgit.NewInspector(run)
	common, err := localgit.ResolveCommonDir(testCtx(t), in.CommonGitDir, repoPath)
	if err != nil {
		t.Fatalf("resolve common dir: %v", err)
	}
	return mutatorOn(t, run, common)
}

// mutatorOn leases common directly (no git call) and returns its mutator.
func mutatorOn(t *testing.T, run cliexec.Exec, common string) *localgit.Mutator {
	t.Helper()
	lease, err := localgit.NewMutationCoordinator(nil).Acquire(testCtx(t), common)
	if err != nil {
		t.Fatalf("acquire lease: %v", err)
	}
	t.Cleanup(func() {
		if err := lease.Release(); err != nil {
			t.Errorf("release: %v", err)
		}
	})
	m, err := lease.Mutator(run)
	if err != nil {
		t.Fatalf("lease mutator: %v", err)
	}
	return m
}
