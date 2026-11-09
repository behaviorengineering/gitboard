package remotegit

import (
	"sort"
	"strings"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/clock"
	"github.com/behaviorengineering/gitboard/pkg/gberr"
)

// staleAfter is how long without updates before a branch is marked stale.
const staleAfter = 14 * 24 * time.Hour

// maxBranchesPerRow caps the branches one project row shows.
const maxBranchesPerRow = 20

type branchAccum struct {
	byName map[string]*board.BranchRef
	order  []string
	now    clock.Clock
}

// newBranchAccum returns an empty accumulator that marks staleness against now.
func newBranchAccum(now clock.Clock) *branchAccum {
	return &branchAccum{
		byName: map[string]*board.BranchRef{},
		now:    now,
	}
}

func (a *branchAccum) ensure(name string) *board.BranchRef {
	name = trimBranch(name)
	if name == "" {
		return nil
	}
	if b, ok := a.byName[name]; ok {
		return b
	}
	b := &board.BranchRef{Name: name}
	a.byName[name] = b
	a.order = append(a.order, name)
	return b
}

func (a *branchAccum) setDefault(name string) {
	if b := a.ensure(name); b != nil {
		b.Default = true
	}
}

// addRemote registers a forge branch head so it appears even without an open review.
// CI history still cannot invent names on its own (see setCI).
func (a *branchAccum) addRemote(name, updatedAt, webURL string) {
	b := a.ensure(name)
	if b == nil {
		return
	}
	if webURL != "" && b.WebURL == "" {
		b.WebURL = webURL
	}
	a.touchUpdated(name, updatedAt)
}

type reviewInfo struct {
	ID        int
	URL       string
	Conflict  bool
	UpdatedAt string
	Draft     bool
}

func (a *branchAccum) setOpenReview(name string, info reviewInfo) {
	b := a.ensure(name)
	if b == nil {
		return
	}
	b.OpenReview = true
	if info.ID > 0 {
		b.ReviewID = info.ID
	}
	if info.URL != "" {
		b.WebURL = info.URL
	}
	if info.Conflict {
		b.Conflict = true
	}
	if info.Draft {
		b.Draft = true
	}
	a.touchUpdated(name, info.UpdatedAt)
}

func (a *branchAccum) setCI(run CachedCIRun) {
	name := trimBranch(run.Branch)
	b, ok := a.byName[name]
	if !ok || b == nil {
		// CI history alone must not invent branches (old Actions refs look like "local" noise).
		return
	}
	// First call wins (callers pass runs newest-first → latest CI per branch).
	if run.Status != "" && b.CIStatus == "" {
		b.CIStatus = run.Status
	}
	if run.URL != "" && b.CIURL == "" {
		b.CIURL = run.URL
	}
	if run.RunID != "" && b.RunID == "" {
		b.RunID = run.RunID
	}
	a.touchUpdated(name, run.UpdatedAt)
}

func (a *branchAccum) touchUpdated(name, updatedAt string) {
	name = trimBranch(name)
	b, ok := a.byName[name]
	if !ok || b == nil || strings.TrimSpace(updatedAt) == "" {
		return
	}
	next, okParse := parseTime(updatedAt)
	if !okParse {
		if b.UpdatedAt == "" {
			b.UpdatedAt = updatedAt
		}
		return
	}
	if b.UpdatedAt == "" {
		b.UpdatedAt = next.UTC().Format(time.RFC3339)
		return
	}
	prev, okPrev := parseTime(b.UpdatedAt)
	if !okPrev || next.After(prev) {
		b.UpdatedAt = next.UTC().Format(time.RFC3339)
	}
}

func (a *branchAccum) list() []board.BranchRef {
	now := a.now()
	out := make([]board.BranchRef, 0, len(a.order))
	for _, name := range a.order {
		bPtr, ok := a.byName[name]
		if !ok || bPtr == nil {
			continue
		}
		b := *bPtr
		if t, ok := parseTime(b.UpdatedAt); ok && now.Sub(t) > staleAfter {
			b.Stale = true
		}
		out = append(out, b)
	}
	sort.SliceStable(out, func(i, j int) bool {
		ai, aj := out[i], out[j]
		if ai.Default != aj.Default {
			return ai.Default
		}
		if ai.Conflict != aj.Conflict {
			return ai.Conflict
		}
		if ai.OpenReview != aj.OpenReview {
			return ai.OpenReview
		}
		ti, oki := parseTime(ai.UpdatedAt)
		tj, okj := parseTime(aj.UpdatedAt)
		if oki && okj && !ti.Equal(tj) {
			return ti.After(tj)
		}
		if oki != okj {
			return oki
		}
		return ai.Name < aj.Name
	})
	if len(out) > maxBranchesPerRow {
		out = out[:maxBranchesPerRow]
	}
	return out
}

func parseTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05 -0700",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func trimBranch(name string) string {
	return strings.TrimSpace(name)
}

// forgeBranchNameOK rejects a branch name that is empty or could be read as
// a CLI flag or argument separator once passed to a forge CLI.
func forgeBranchNameOK(branch string) error {
	const op = "remotegit.forgeBranchNameOK"
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return gberr.New(gberr.CodeInvalid, op, "branch is required")
	}
	if strings.ContainsAny(branch, " \t\n:") || strings.HasPrefix(branch, "-") {
		return gberr.New(gberr.CodeInvalid, op, "branch name is not valid").With("branch", branch)
	}
	return nil
}
