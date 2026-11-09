package localgit

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// refLines joins for-each-ref rows into command output.
func refLines(rows ...string) string {
	return strings.Join(rows, "\n") + "\n"
}

func TestParseTrack(t *testing.T) {
	cases := []struct {
		in           string
		ahead, behin int
		gone, bad    bool
	}{
		{in: ""},
		{in: "ahead 3", ahead: 3},
		{in: "behind 2", behin: 2},
		{in: "ahead 3, behind 2", ahead: 3, behin: 2},
		{in: "gone", gone: true},
		{in: "ahead x", bad: true},
		{in: "sideways 1", bad: true},
	}
	for _, tc := range cases {
		ahead, behind, err := parseTrack(tc.in)
		switch {
		case tc.gone:
			if !errors.Is(err, errUpstreamGone) {
				t.Errorf("%q: want errUpstreamGone, got %v", tc.in, err)
			}
		case tc.bad:
			if err == nil {
				t.Errorf("%q: want error", tc.in)
			}
		case err != nil || ahead != tc.ahead || behind != tc.behin:
			t.Errorf("%q: ahead=%d behind=%d err=%v", tc.in, ahead, behind, err)
		}
	}
}

func TestParseOriginRefs(t *testing.T) {
	refs, err := parseOriginRefs([]byte(refLines(
		"refs/heads/tracked\x00refs/remotes/origin/tracked\x00ahead 1, behind 4\x00",
		"refs/heads/no-upstream\x00\x00\x00",
		"refs/heads/gone\x00refs/remotes/origin/gone\x00gone\x00",
		"refs/heads/other-up\x00refs/remotes/origin/elsewhere\x00behind 1\x00",
		"refs/remotes/origin/HEAD\x00\x00\x00refs/remotes/origin/trunk",
		"refs/remotes/origin/tracked\x00\x00\x00",
		"refs/remotes/origin/no-upstream\x00\x00\x00",
		"refs/remotes/origin/other-up\x00\x00\x00",
		"refs/remotes/origin/trunk\x00\x00\x00",
	)))
	if err != nil {
		t.Fatal(err)
	}
	if refs.originHead != "trunk" || refs.defaultBranch() != "trunk" {
		t.Fatalf("origin HEAD=%q default=%q", refs.originHead, refs.defaultBranch())
	}
	if len(refs.heads) != 4 {
		t.Fatalf("heads=%+v", refs.heads)
	}
	if _, ok := refs.remote["HEAD"]; ok {
		t.Fatal("origin/HEAD must not count as a branch")
	}
	want := map[string]bool{"tracked": true, "no-upstream": false, "gone": false, "other-up": false}
	for _, h := range refs.heads {
		s, ok := h.listedSync(refs.originHead)
		if ok != want[h.name] {
			t.Errorf("%s: listed=%v want %v", h.name, ok, want[h.name])
		}
		if h.name == "tracked" && (s.Ahead != 1 || s.Behind != 4) {
			t.Errorf("tracked sync=%+v", s)
		}
	}
}

func TestOriginRefsDefaultFallback(t *testing.T) {
	refs := originRefs{remote: map[string]struct{}{"master": {}}}
	if got := refs.defaultBranch(); got != "master" {
		t.Fatalf("default=%q", got)
	}
	refs.remote["main"] = struct{}{}
	if got := refs.defaultBranch(); got != "main" {
		t.Fatalf("default=%q", got)
	}
	if got := (originRefs{}).defaultBranch(); got != "" {
		t.Fatalf("default=%q", got)
	}
}

// TestOriginSyncFallsBackOnlyForUntrackedTwins runs rev-list for the branch
// whose upstream is not its origin twin and for no other.
func TestOriginSyncFallsBackOnlyForUntrackedTwins(t *testing.T) {
	fk := &fakeExec{responses: map[string][]byte{
		"for-each-ref --format=" + refFormat + " refs/heads/": []byte(refLines(
			"refs/heads/main\x00refs/remotes/origin/main\x00\x00",
			"refs/heads/tracked\x00refs/remotes/origin/tracked\x00ahead 2\x00",
			"refs/heads/loose\x00\x00\x00",
			"refs/heads/local-only\x00\x00\x00",
			"refs/heads/gone\x00refs/remotes/origin/gone\x00gone\x00",
			"refs/remotes/origin/HEAD\x00\x00\x00refs/remotes/origin/main",
			"refs/remotes/origin/main\x00\x00\x00",
			"refs/remotes/origin/tracked\x00\x00\x00",
			"refs/remotes/origin/loose\x00\x00\x00",
		)),
		"rev-list --left-right --count refs/remotes/origin/loose...refs/heads/loose": []byte("3\t1\n"),
	}}
	rs, err := NewInspector(fk).OriginSync(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("OriginSync: %v calls=%v", err, fk.calls)
	}
	got := map[string]BranchSync{}
	for _, b := range rs.Branches {
		got[b.Name] = b
	}
	if len(got) != 3 || got["tracked"].Ahead != 2 || got["loose"].Ahead != 1 || got["loose"].Behind != 3 {
		t.Fatalf("branches=%+v", rs.Branches)
	}
	if n := countCalls(fk, "rev-list"); n != 1 {
		t.Fatalf("rev-list calls=%d want 1: %v", n, fk.calls)
	}
}

// TestOriginSyncUsesAheadBehindOnNewGit reads the default branch position
// from %(ahead-behind:...) when git is 2.41 or newer, so the default branch
// needs no rev-list even without an upstream.
func TestOriginSyncUsesAheadBehindOnNewGit(t *testing.T) {
	fk := &fakeExec{responses: map[string][]byte{
		"-C . version": []byte("git version 2.46.2\n"),
		"for-each-ref --format=" + refFormat + aheadBehindAtom: []byte(refLines(
			"refs/heads/main\x00\x00\x00\x002 5",
			"refs/remotes/origin/HEAD\x00\x00\x00refs/remotes/origin/main\x00",
			"refs/remotes/origin/main\x00\x00\x00\x000 0",
		)),
	}}
	in := NewInspector(fk)
	st := Status{Mapped: true, Path: t.TempDir()}
	in.EnrichOriginSync(context.Background(), &st)
	if st.Error != "" || st.DefaultAhead != 2 || st.DefaultBehind != 5 {
		t.Fatalf("status=%+v calls=%v", st, fk.calls)
	}
	in.EnrichOriginSync(context.Background(), &st)
	if n := countCalls(fk, "version"); n != 1 {
		t.Fatalf("git version probed %d times, want once", n)
	}
	if n := countCalls(fk, "rev-list"); n != 0 {
		t.Fatalf("rev-list calls=%d", n)
	}
}

// TestOriginSyncRetriesWithoutAtomWhenOriginHEADMissing lists again without
// %(ahead-behind:...) when git rejects it because origin/HEAD is unset.
func TestOriginSyncRetriesWithoutAtomWhenOriginHEADMissing(t *testing.T) {
	fk := &fakeExec{
		responses: map[string][]byte{
			"-C . version": []byte("git version 2.46.2\n"),
			"for-each-ref --format=" + refFormat + " refs/heads/": []byte(refLines(
				"refs/heads/main\x00refs/remotes/origin/main\x00behind 1\x00",
				"refs/remotes/origin/main\x00\x00\x00",
			)),
		},
		errors: map[string]error{
			"for-each-ref --format=" + refFormat + aheadBehindAtom: fmtError("fatal: failed to find 'refs/remotes/origin/HEAD'"),
		},
	}
	rs, err := NewInspector(fk).OriginSync(context.Background(), t.TempDir())
	if err != nil || rs.DefaultBranch != "main" || len(rs.Branches) != 1 || rs.Branches[0].Behind != 1 {
		t.Fatalf("rs=%+v err=%v", rs, err)
	}
}

func TestOriginSyncOldGitSkipsAtom(t *testing.T) {
	fk := &fakeExec{responses: map[string][]byte{
		"-C . version": []byte("git version 2.39.3 (Apple Git-146)\n"),
		"for-each-ref --format=" + refFormat + " refs/heads/": []byte(refLines(
			"refs/heads/main\x00refs/remotes/origin/main\x00\x00",
			"refs/remotes/origin/main\x00\x00\x00",
		)),
	}}
	if _, err := NewInspector(fk).OriginSync(context.Background(), t.TempDir()); err != nil {
		t.Fatalf("OriginSync: %v calls=%v", err, fk.calls)
	}
	if n := countCalls(fk, "ahead-behind"); n != 0 {
		t.Fatalf("old git must not get the atom: %v", fk.calls)
	}
}

func TestOriginSyncCompareFailureKeepsDefault(t *testing.T) {
	fk := &fakeExec{
		responses: map[string][]byte{
			"for-each-ref --format=" + refFormat + " refs/heads/": []byte(refLines(
				"refs/heads/loose\x00\x00\x00",
				"refs/remotes/origin/HEAD\x00\x00\x00refs/remotes/origin/main",
				"refs/remotes/origin/loose\x00\x00\x00",
			)),
		},
		errors: map[string]error{"rev-list": fmtError("boom")},
	}
	in := NewInspector(fk)
	st := Status{Mapped: true, Path: t.TempDir()}
	in.EnrichOriginSync(context.Background(), &st)
	if st.DefaultBranch != "main" || st.OriginSync != nil || !strings.Contains(st.Error, "could not compare local branches to origin") {
		t.Fatalf("status=%+v", st)
	}
}

func TestParseGitVersion(t *testing.T) {
	cases := map[string]bool{
		"git version 2.46.2":                 true,
		"git version 2.41.0":                 true,
		"git version 2.40.1":                 false,
		"git version 2.39.3 (Apple Git-146)": false,
		"git version 3.0.0":                  true,
	}
	for in, want := range cases {
		v, ok := parseGitVersion(in)
		if !ok || v.atLeast(minAheadBehind) != want {
			t.Errorf("%q: v=%+v ok=%v", in, v, ok)
		}
	}
	if _, ok := parseGitVersion("not git"); ok {
		t.Error("garbage parsed")
	}
}

func countCalls(f *fakeExec, substr string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if strings.Contains(c, substr) {
			n++
		}
	}
	return n
}
