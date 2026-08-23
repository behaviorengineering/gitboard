package forge

import (
	"sort"
	"strings"
	"time"
)

// StaleAfter is how long without updates before a branch is marked stale.
const StaleAfter = 14 * 24 * time.Hour

type branchAccum struct {
	byName map[string]*BranchRef
	order  []string
}

func newBranchAccum() *branchAccum {
	return &branchAccum{byName: map[string]*BranchRef{}}
}

func (a *branchAccum) ensure(name string) *BranchRef {
	name = trimBranch(name)
	if name == "" {
		return nil
	}
	if b, ok := a.byName[name]; ok {
		return b
	}
	b := &BranchRef{Name: name}
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

func (a *branchAccum) setCI(name, status, webURL, updatedAt, runID string) {
	name = trimBranch(name)
	b, ok := a.byName[name]
	if !ok || b == nil {
		// CI history alone must not invent branches (old Actions refs look like "local" noise).
		return
	}
	// First call wins (callers pass runs newest-first → latest CI per branch).
	if status != "" && b.CIStatus == "" {
		b.CIStatus = status
	}
	if webURL != "" && b.CIURL == "" {
		b.CIURL = webURL
	}
	if runID != "" && b.RunID == "" {
		b.RunID = runID
	}
	a.touchUpdated(name, updatedAt)
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

func (a *branchAccum) list() []BranchRef {
	now := time.Now().UTC()
	out := make([]BranchRef, 0, len(a.order))
	for _, name := range a.order {
		b := *a.byName[name]
		if t, ok := parseTime(b.UpdatedAt); ok && now.Sub(t) > StaleAfter {
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
	const max = 20
	if len(out) > max {
		out = out[:max]
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

func githubHasConflict(mergeable, mergeState string) bool {
	m := strings.ToUpper(strings.TrimSpace(mergeable))
	s := strings.ToUpper(strings.TrimSpace(mergeState))
	return m == "CONFLICTING" || s == "DIRTY" || s == "CONFLICTING"
}

func gitlabHasConflict(hasConflicts bool, mergeStatus string) bool {
	if hasConflicts {
		return true
	}
	s := strings.ToLower(strings.TrimSpace(mergeStatus))
	return s == "cannot_be_merged" || s == "cannot_be_merged_recheck"
}
