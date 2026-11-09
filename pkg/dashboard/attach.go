package dashboard

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
)

// Worker limits. Each has its own budget: projects run collectParallel at
// once, and each project inspects at most inspectParallel checkouts and
// fetches at most originFetchParallel repositories at once.
const (
	collectParallel     = 8
	inspectParallel     = 4
	originFetchParallel = 8
)

// localInputs is what attachLocal needs besides the project.
type localInputs struct {
	disc          localgit.Discovery
	labels        map[string]string // forge project label by localgit.TrackKey.
	fresh         bool
	fetchTTL      time.Duration
	forgeBranches []string
}

// inspectTarget is one checkout to inspect and the key of its repository.
type inspectTarget struct {
	checkout localgit.Checkout
	// common is localgit.CommonKey: it keys fetch errors, mutation epochs and
	// the shared origin sync, so writers and readers agree.
	common    string
	commonErr error
}

// attachLocal inspects every checkout of p: it refreshes origin once per
// repository, inspects each checkout, shares one origin sync per repository
// across its checkouts, and builds the board's local status.
func (s *Service) attachLocal(ctx context.Context, p boardconfig.Project, in localInputs) *board.LocalStatus {
	if s.local == nil {
		return nil
	}
	primary, list, early, ok := resolveInspectList(p, in.disc)
	if !ok {
		return early
	}
	targets := s.inspectTargets(ctx, list)
	policy := localgit.FetchPolicy{
		TTL:      in.fetchTTL,
		Fresh:    in.fresh,
		Branches: s.originFetchBranches(ctx, list, in.forgeBranches),
	}
	fetchErr := s.refreshOrigins(ctx, targets, policy)
	// Capture epochs after our own fetch bump so only concurrent mutations invalidate.
	epochs := s.mutationEpochs(targets)
	statuses := s.inspectAll(ctx, targets, fetchErr, epochs)
	return s.buildLocalStatus(ctx, primary, targets, statuses, fetchErr, in.labels)
}

// resolveInspectList picks the primary checkout and the list to inspect.
// The primary is always in the list, even when an explicit local_path is
// outside the scan roots. ok is false when there is nothing to inspect;
// early is then the row's local status (nil when no roots were scanned).
func resolveInspectList(p boardconfig.Project, disc localgit.Discovery) (primary localgit.Checkout, list []localgit.Checkout, early *board.LocalStatus, ok bool) {
	checkouts := disc.Appearances(p.Host, p.Path)
	explicit := strings.TrimSpace(p.LocalPath)
	if len(checkouts) == 0 && explicit == "" {
		if len(disc.ByKey) == 0 {
			return localgit.Checkout{}, nil, nil, false
		}
		return localgit.Checkout{}, nil, &board.LocalStatus{Mapped: false}, false
	}
	primary, ok = localgit.PickPrimary(explicit, checkouts)
	if !ok {
		return localgit.Checkout{}, nil, &board.LocalStatus{Mapped: false}, false
	}
	if explicit == "" {
		return primary, checkouts, nil, true
	}
	if expanded, err := localgit.ExpandPath(primary.Path); err == nil {
		primary.Path = expanded
	}
	for _, c := range checkouts {
		if filepath.Clean(c.Path) == filepath.Clean(primary.Path) {
			return primary, checkouts, nil, true
		}
	}
	localgit.FillCheckoutMeta(&primary)
	return primary, append([]localgit.Checkout{primary}, checkouts...), nil, true
}

// inspectTargets resolves each checkout's repository key once.
func (s *Service) inspectTargets(ctx context.Context, list []localgit.Checkout) []inspectTarget {
	out := make([]inspectTarget, len(list))
	for i, c := range list {
		key, err := localgit.CommonKey(ctx, s.local.CommonGitDir, c)
		out[i] = inspectTarget{checkout: c, common: key, commonErr: err}
	}
	return out
}

// refreshOrigins runs a TTL-gated origin refresh once per repository, at
// most originFetchParallel at once. Fresh or expired full TTL uses
// `git fetch --prune origin`; a warm TTL fetches only policy.Branches (none
// skips the network). It returns failures keyed by inspectTarget.common.
func (s *Service) refreshOrigins(ctx context.Context, targets []inspectTarget, policy localgit.FetchPolicy) map[string]error {
	out := map[string]error{}
	var jobs []inspectTarget
	seen := map[string]struct{}{}
	for _, t := range targets {
		if t.common == "" {
			continue
		}
		if _, ok := seen[t.common]; ok {
			continue
		}
		seen[t.common] = struct{}{}
		if t.commonErr != nil {
			out[t.common] = t.commonErr
			continue
		}
		jobs = append(jobs, t)
	}
	errs := make([]error, len(jobs))
	parallel(len(jobs), originFetchParallel, func(i int) {
		errs[i] = s.fetchOriginLocked(ctx, jobs[i].checkout.Path, jobs[i].common, policy)
	})
	for i, err := range errs {
		if err != nil {
			out[jobs[i].common] = err
		}
	}
	return out
}

func (s *Service) fetchOriginLocked(ctx context.Context, path, common string, policy localgit.FetchPolicy) error {
	return s.withLease(ctx, common, func(m LocalMutator) (bool, error) {
		return m.FetchOriginSmart(ctx, path, policy, s.originFetch)
	})
}

// mutationEpochs records each resolved repository's mutation epoch.
func (s *Service) mutationEpochs(targets []inspectTarget) map[string]uint64 {
	out := map[string]uint64{}
	for _, t := range targets {
		if t.commonErr != nil || t.common == "" {
			continue
		}
		if _, ok := out[t.common]; !ok {
			out[t.common] = s.mutations.Epoch(t.common)
		}
	}
	return out
}

// mutationEpochChanged reports whether a mutation on t's repository finished
// since before was taken.
func (s *Service) mutationEpochChanged(t inspectTarget, before map[string]uint64) bool {
	prev, ok := before[t.common]
	return ok && s.mutations.Epoch(t.common) != prev
}

// inspectAll inspects each target, at most inspectParallel at once, then
// applies one origin sync per repository, the repository's fetch error, and
// drops origin sync when a concurrent mutation made it stale.
func (s *Service) inspectAll(ctx context.Context, targets []inspectTarget, fetchErr map[string]error, epochs map[string]uint64) []localgit.Status {
	statuses := make([]localgit.Status, len(targets))
	parallel(len(targets), inspectParallel, func(i int) {
		statuses[i] = s.local.InspectPath(ctx, targets[i].checkout.Path)
	})
	syncs := s.originSyncs(ctx, targets, statuses)
	for i, t := range targets {
		st := &statuses[i]
		if r, ok := syncs[t.common]; ok && syncable(*st) {
			r.apply(st)
		}
		if err := fetchErr[t.common]; err != nil {
			st.Error = appendErrText(st.Error, err)
			localgit.InvalidateOriginSync(st)
		}
		if s.mutationEpochChanged(t, epochs) {
			localgit.InvalidateOriginSync(st)
		}
	}
	return statuses
}

// repoSync is one repository's origin sync and its failure.
type repoSync struct {
	sync localgit.RepoSync
	err  error
}

func (r repoSync) apply(st *localgit.Status) {
	r.sync.Apply(st)
	if r.err != nil {
		st.Error = appendErrText(st.Error, r.err)
		localgit.InvalidateOriginSync(st)
	}
}

// syncable reports whether st inspected cleanly, so origin sync applies.
func syncable(st localgit.Status) bool {
	return st.Mapped && st.Path != "" && st.Error == ""
}

// originSyncs computes origin sync once per repository, from the first
// checkout of it that inspected cleanly. Refs live in the common git dir, so
// the result holds for every worktree of that repository.
func (s *Service) originSyncs(ctx context.Context, targets []inspectTarget, statuses []localgit.Status) map[string]repoSync {
	type job struct{ key, path string }
	var jobs []job
	seen := map[string]struct{}{}
	for i, t := range targets {
		if t.common == "" || !syncable(statuses[i]) {
			continue
		}
		if _, ok := seen[t.common]; ok {
			continue
		}
		seen[t.common] = struct{}{}
		jobs = append(jobs, job{key: t.common, path: statuses[i].Path})
	}
	results := make([]repoSync, len(jobs))
	parallel(len(jobs), inspectParallel, func(i int) {
		rs, err := s.local.OriginSync(ctx, jobs[i].path)
		results[i] = repoSync{sync: rs, err: err}
	})
	out := make(map[string]repoSync, len(jobs))
	for i, j := range jobs {
		out[j.key] = results[i]
	}
	return out
}

// buildLocalStatus turns inspected checkouts into the board's local status:
// one appearance per checkout, the union of their worktrees, and the
// primary checkout's fields at the top level.
func (s *Service) buildLocalStatus(ctx context.Context, primary localgit.Checkout, targets []inspectTarget, statuses []localgit.Status, fetchErr map[string]error, labels map[string]string) *board.LocalStatus {
	parentLabels := map[string]string{}
	appearances := make([]board.LocalAppearance, 0, len(targets))
	var union []board.LocalWorktree
	var primaryLocal *board.LocalStatus
	primaryPath := filepath.Clean(primary.Path)
	for i, t := range targets {
		c := t.checkout
		parentLabel := ""
		if c.Role == board.AppearanceSubmodule && c.Superproject != "" {
			parentLabel = s.parentLabel(ctx, c.Superproject, labels, parentLabels)
		}
		displayID := localgit.DisplayID(c, parentLabel)
		app := statusToAppearance(c, statuses[i], displayID, parentLabel)
		app.Primary = filepath.Clean(c.Path) == primaryPath
		appearances = append(appearances, app)
		union = append(union, app.Worktrees...)
		if app.Primary {
			primaryLocal = toBoardLocal(statuses[i])
		}
	}
	if primaryLocal == nil {
		// resolveInspectList always lists the primary; this only guards a
		// path spelling that did not compare equal.
		one := s.inspectTargets(ctx, []localgit.Checkout{primary})
		primaryLocal = toBoardLocal(s.inspectAll(ctx, one, fetchErr, nil)[0])
	}
	primaryLocal.Appearances = appearances
	primaryLocal.Worktrees = union
	return primaryLocal
}

// originFetchBranches unions visible forge branch names with local heads so
// OriginSync and attention ranking stay accurate after a targeted fetch.
func (s *Service) originFetchBranches(ctx context.Context, checkouts []localgit.Checkout, forgeBranches []string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	for _, name := range forgeBranches {
		add(name)
	}
	for _, c := range checkouts {
		if c.Path == "" || ctx.Err() != nil {
			continue
		}
		heads, err := s.local.ListLocalHeads(ctx, c.Path)
		if err != nil {
			continue
		}
		for _, name := range heads {
			add(name)
		}
	}
	return out
}

func (s *Service) parentLabel(ctx context.Context, parentPath string, labelByKey map[string]string, cache map[string]string) string {
	parentPath = filepath.Clean(parentPath)
	if v, ok := cache[parentPath]; ok {
		return v
	}
	label := filepath.Base(parentPath)
	if url, err := s.local.OriginRemote(ctx, parentPath); err == nil {
		if ref, ok := localgit.ParseRemoteURL(url); ok {
			if lb, ok := labelByKey[localgit.TrackKey(ref.Host, ref.Path)]; ok {
				label = lb
			}
		}
	}
	cache[parentPath] = label
	return label
}

func statusToAppearance(c localgit.Checkout, st localgit.Status, displayID, parentLabel string) board.LocalAppearance {
	app := board.LocalAppearance{
		Role:          c.Role,
		Path:          c.Path,
		DisplayID:     displayID,
		ParentPath:    c.Superproject,
		ParentLabel:   parentLabel,
		RelPath:       c.RelPath,
		Error:         st.Error,
		Branch:        st.Branch,
		Tag:           st.Tag,
		Detached:      st.Detached,
		Dirty:         st.Dirty,
		Ahead:         st.Ahead,
		Behind:        st.Behind,
		Upstream:      st.Upstream,
		DefaultBranch: st.DefaultBranch,
		DefaultBehind: st.DefaultBehind,
		DefaultAhead:  st.DefaultAhead,
		OriginSync:    originSyncToBoard(st.OriginSync),
	}
	if app.Role == "" {
		app.Role = board.AppearanceStandalone
	}
	for _, wt := range boardWorktrees(st.Worktrees) {
		wt.AppearancePath = c.Path
		wt.AppearanceLabel = displayID
		app.Worktrees = append(app.Worktrees, wt)
	}
	return app
}

func toBoardLocal(st localgit.Status) *board.LocalStatus {
	return &board.LocalStatus{
		Mapped:        st.Mapped,
		Path:          st.Path,
		Error:         st.Error,
		Branch:        st.Branch,
		Tag:           st.Tag,
		Detached:      st.Detached,
		Dirty:         st.Dirty,
		Ahead:         st.Ahead,
		Behind:        st.Behind,
		Upstream:      st.Upstream,
		DefaultBranch: st.DefaultBranch,
		DefaultBehind: st.DefaultBehind,
		DefaultAhead:  st.DefaultAhead,
		OriginSync:    originSyncToBoard(st.OriginSync),
		Worktrees:     boardWorktrees(st.Worktrees),
	}
}

// boardWorktrees converts non-bare worktrees to board rows.
func boardWorktrees(trees []localgit.Worktree) []board.LocalWorktree {
	var out []board.LocalWorktree
	for _, wt := range trees {
		if wt.Bare {
			continue
		}
		out = append(out, board.LocalWorktree{
			Path:     wt.Path,
			Branch:   wt.Branch,
			Tag:      wt.Tag,
			Detached: wt.Detached,
			Bare:     wt.Bare,
			Main:     wt.Main,
			Dirty:    wt.Dirty,
			Ahead:    wt.Ahead,
			Behind:   wt.Behind,
			Upstream: wt.Upstream,
		})
	}
	return out
}

func originSyncToBoard(in []localgit.BranchSync) []board.BranchOriginSync {
	if len(in) == 0 {
		return nil
	}
	out := make([]board.BranchOriginSync, len(in))
	for i, s := range in {
		out[i] = board.BranchOriginSync{Name: s.Name, Ahead: s.Ahead, Behind: s.Behind}
	}
	return out
}

// parallel runs fn(i) for every i below n, at most limit at once, and waits
// for all of them. fn reports through its own result slots.
func parallel(n, limit int, fn func(i int)) {
	if limit < 1 {
		limit = 1
	}
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i := range n {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			fn(i)
		}()
	}
	wg.Wait()
}
