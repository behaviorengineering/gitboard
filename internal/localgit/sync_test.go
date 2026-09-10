package localgit

import (
	"context"
	"strings"
	"testing"
)

func TestInspectSyncStates(t *testing.T) {
	tests := []struct {
		name         string
		count        string
		status       string
		mergeBaseErr bool
		wantRelation string
		wantAhead    int
		wantBehind   int
		wantFast     bool
	}{
		{
			name:         "up to date",
			count:        "0\t0\n",
			wantRelation: "up_to_date",
		},
		{
			name:         "behind only",
			count:        "2\t0\n",
			wantRelation: "behind_only",
			wantBehind:   2,
			wantFast:     true,
		},
		{
			name:         "ahead only",
			count:        "0\t1\n",
			wantRelation: "ahead_only",
			wantAhead:    1,
		},
		{
			name:         "diverged",
			count:        "2\t1\n",
			wantRelation: "diverged",
			wantAhead:    1,
			wantBehind:   2,
		},
		{
			name:         "dirty behind only",
			count:        "2\t0\n",
			status:       " M config.yaml\n?? notes.txt\n",
			wantRelation: "behind_only",
			wantBehind:   2,
		},
		{
			name:         "unrelated",
			count:        "1\t1\n",
			mergeBaseErr: true,
			wantRelation: "unrelated",
			wantAhead:    1,
			wantBehind:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			fx := &fakeExec{
				responses: map[string][]byte{
					"rev-parse --git-dir":                         []byte(".git\n"),
					"fetch --prune origin":                        nil,
					"rev-parse --verify refs/heads/main":          []byte("local-sha\n"),
					"rev-parse --verify refs/remotes/origin/main": []byte("remote-sha\n"),
					"rev-parse --abbrev-ref HEAD":                 []byte("main\n"),
					"status --porcelain":                          []byte(tt.status),
					"rev-list --left-right --count refs/remotes/origin/main...refs/heads/main":       []byte(tt.count),
					"merge-base refs/heads/main refs/remotes/origin/main":                            []byte("base-sha\n"),
					"log --max-count 20 --format=%h%x09%s refs/heads/main..refs/remotes/origin/main": []byte("r1\tremote change\n"),
					"log --max-count 20 --format=%h%x09%s refs/remotes/origin/main..refs/heads/main": []byte("l1\tlocal change\n"),
				},
			}
			if tt.mergeBaseErr {
				fx.errors = map[string]error{
					"merge-base refs/heads/main refs/remotes/origin/main": fmtError("unrelated"),
				}
			}

			in := NewInspector(fx)
			got, err := in.InspectSync(context.Background(), dir, "main")
			if err != nil {
				t.Fatal(err)
			}
			if got.Relation != tt.wantRelation {
				t.Fatalf("relation=%q want %q", got.Relation, tt.wantRelation)
			}
			if got.AheadCount != tt.wantAhead || got.BehindCount != tt.wantBehind {
				t.Fatalf("counts ahead=%d behind=%d want ahead=%d behind=%d", got.AheadCount, got.BehindCount, tt.wantAhead, tt.wantBehind)
			}
			if got.CanFastForward != tt.wantFast {
				t.Fatalf("can fast-forward=%v want %v", got.CanFastForward, tt.wantFast)
			}
			if tt.status != "" && len(got.DirtyFiles) != 2 {
				t.Fatalf("dirty files=%v", got.DirtyFiles)
			}
			if tt.mergeBaseErr && len(got.Ahead) != 0 {
				t.Fatalf("unrelated comparison should not list commits: %+v", got.Ahead)
			}
		})
	}
}

func TestInspectSyncValidatesBranch(t *testing.T) {
	in := NewInspector(&fakeExec{})
	_, err := in.InspectSync(context.Background(), t.TempDir(), "bad branch")
	if err == nil || !strings.Contains(err.Error(), "invalid branch") {
		t.Fatalf("want invalid branch error, got %v", err)
	}
}
