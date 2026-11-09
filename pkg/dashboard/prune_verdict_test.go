package dashboard

import (
	"context"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// safeInputs is a branch every rule allows: gone from the remote, clean,
// nothing open, and a merged review on the forge.
func safeInputs() PruneInputs {
	return PruneInputs{
		Branch:        "feat/gone",
		DefaultBranch: "main",
		RemoteNamesOK: true,
		OpenOK:        true,
		Merged:        true,
		MergedOK:      true,
		MergedChecked: true,
	}
}

func TestPruneVerdict(t *testing.T) {
	cases := []struct {
		name   string
		edit   func(*PruneInputs)
		allow  bool
		hint   board.PruneHint
		cand   bool
		reason string
	}{
		{name: "merged review allows", edit: func(*PruneInputs) {}, allow: true, hint: board.PruneSafe, cand: true},
		{name: "content on default allows", edit: func(in *PruneInputs) { in.Merged, in.ContentOnDefault = false, true }, allow: true, hint: board.PruneSafe, cand: true},
		{name: "remote names failed", edit: func(in *PruneInputs) { in.RemoteNamesOK = false }, reason: reasonRemoteUnknown},
		{name: "open reviews failed", edit: func(in *PruneInputs) { in.OpenOK = false }, reason: reasonOpenUnknown},
		{name: "open review", edit: func(in *PruneInputs) { in.OpenReview = true }, reason: reasonOpenReview},
		{name: "still on remote", edit: func(in *PruneInputs) { in.OnRemote = true }, reason: reasonOnRemote},
		{name: "dirty", edit: func(in *PruneInputs) { in.Dirty = true }, reason: reasonDirty},
		{name: "default branch", edit: func(in *PruneInputs) { in.Branch = "main" }, reason: reasonDefault},
		{name: "default unknown", edit: func(in *PruneInputs) { in.DefaultBranch = "" }, reason: reasonDefaultUnknown},
		{name: "detached or bare", edit: func(in *PruneInputs) { in.NotBranch = true }, reason: reasonNoBranch},
		{name: "no branch", edit: func(in *PruneInputs) { in.Branch = " " }, reason: reasonNoBranch},
		{name: "merged lookup failed", edit: func(in *PruneInputs) { in.Merged, in.MergedChecked = false, false }, cand: true, reason: reasonMergedUnknown},
		{name: "bulk merged list failed", edit: func(in *PruneInputs) { in.Merged, in.MergedOK = false, false }, cand: true, reason: reasonMergedUnknown},
		{name: "content not on default is only likely", edit: func(in *PruneInputs) { in.Merged = false }, hint: board.PruneLikely, cand: true, reason: reasonUnproven},
		{name: "open review beats merged and content", edit: func(in *PruneInputs) { in.OpenReview, in.ContentOnDefault = true, true }, reason: reasonOpenReview},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := safeInputs()
			tc.edit(&in)
			got := pruneVerdict(in)
			want := PruneVerdict{Allow: tc.allow, Hint: tc.hint, Candidate: tc.cand, Reason: tc.reason}
			if got != want {
				t.Fatalf("verdict=%+v want %+v", got, want)
			}
		})
	}
}

// pruneLocalFake is one clean checkout on feat/gone whose writer counts
// RemoveSafeCheckout calls.
type pruneLocalFake struct {
	pullLocalFake
	removes atomic.Int32
}

func (f *pruneLocalFake) InspectPath(_ context.Context, path string) localgit.Status {
	return localgit.Status{
		Mapped: true, Path: path, Branch: "feat/gone",
		Worktrees: []localgit.Worktree{{Path: path, Branch: "feat/gone", Main: true}},
	}
}

func (f *pruneLocalFake) OriginSync(context.Context, string) (localgit.RepoSync, error) {
	return localgit.RepoSync{DefaultBranch: "main"}, nil
}

func (f *pruneLocalFake) RemoveSafeCheckout(context.Context, string, string, string) error {
	f.removes.Add(1)
	return nil
}

func (f *pruneLocalFake) Mutator(*localgit.MutationLease) (LocalMutator, error) {
	return f, nil
}

// TestPruneSafeRemovesOnlyOnAllow checks that PruneSafe reaches the
// Mutator's RemoveSafeCheckout when pruneVerdict allows, and never when it
// refuses.
func TestPruneSafeRemovesOnlyOnAllow(t *testing.T) {
	merged := []byte(`[{"number":9,"headRefName":"feat/gone","url":"https://example/pr/9","mergedAt":"2026-01-01T12:00:00Z"}]`)
	cases := []struct {
		name     string
		open     bool // open reviews load.
		merged   []byte
		removes  int32
		wantCode gberr.Code
		wantMsg  string
	}{
		{name: "merged and nothing open", open: true, merged: merged, removes: 1},
		{name: "open reviews did not load", open: false, merged: merged, wantCode: gberr.CodeConflict, wantMsg: "open reviews"},
		{name: "no merged review", open: true, merged: []byte(`[]`), wantCode: gberr.CodeConflict, wantMsg: reasonUnproven},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Clean(t.TempDir())
			fk := pruneForgeFake(tc.open, tc.merged)
			local := &pruneLocalFake{}
			cmds := NewCommands(newTestService(fk, local))
			doc := boardconfig.File{Projects: []boardconfig.Project{{
				ID: "gh-app", Host: boardconfig.HostGitHub, Path: "acme/app", LocalPath: path,
			}}}
			err := cmds.PruneSafe(t.Context(), doc, PruneSafeRequest{ProjectID: "gh-app", Branch: "feat/gone", WorktreePath: path})
			if got := local.removes.Load(); got != tc.removes {
				t.Fatalf("RemoveSafeCheckout calls=%d want %d (err=%v)", got, tc.removes, err)
			}
			if tc.wantCode == "" && err != nil {
				t.Fatalf("allow: %v", err)
			}
			if tc.wantCode == "" {
				return
			}
			if !gberr.HasCode(err, tc.wantCode) {
				t.Fatalf("code=%q want %q: %v", gberr.CodeOf(err), tc.wantCode, err)
			}
			e, _ := gberr.As(err)
			if !strings.Contains(e.Msg, tc.wantMsg) || strings.Contains(e.Msg, path) {
				t.Fatalf("message %q: want %q and no path", e.Msg, tc.wantMsg)
			}
		})
	}
}

// pruneForgeFake is a GitHub forge for acme/app with only main on the
// remote. open says whether the open-review list loads; merged is the bulk
// merged list.
func pruneForgeFake(open bool, merged []byte) *fakeExec {
	fk := &fakeExec{responses: map[string][]byte{
		"auth status":             []byte(""),
		"repos/acme/app --jq":     []byte(`{"default":"main"}`),
		"repos/acme/app/branches": []byte(`[{"name":"main","commit":{"commit":{"committer":{"date":"2026-01-01T00:00:00Z"}}}}]`),
		"run list":                []byte(`[]`),
		"--state merged":          merged,
		"--head feat/gone":        []byte(`[]`),
	}}
	if open {
		fk.responses["--state open"] = []byte(`[]`)
	}
	return fk
}
