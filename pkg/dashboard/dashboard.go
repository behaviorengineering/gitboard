package dashboard

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// Service aggregates project rows via forge CLIs and optional local git.
type Service struct {
	GitHub      *remotegit.GitHub
	GitLab      *remotegit.GitLab
	AzureDevOps *remotegit.AzureDevOps
	Bitbucket   *remotegit.Bitbucket
	Local       LocalGit
	Cache       *remotegit.TTLCache
	OriginFetch *localgit.OriginFetchCache
	Mutations   *localgit.MutationCoordinator

	scanMu   sync.Mutex
	scanKey  string
	scanAt   time.Time
	scanDisc localgit.Discovery
}

// New returns a dashboard service with in-memory upstream and origin-fetch caches.
func New(gh *remotegit.GitHub, gl *remotegit.GitLab, az *remotegit.AzureDevOps, bb *remotegit.Bitbucket, local LocalGit) *Service {
	return &Service{
		GitHub:      gh,
		GitLab:      gl,
		AzureDevOps: az,
		Bitbucket:   bb,
		Local:       local,
		Cache:       remotegit.NewTTLCache(),
		OriginFetch: localgit.NewOriginFetchCache(),
		Mutations:   localgit.NewMutationCoordinator(localgit.DefaultOSLocker()),
	}
}

// ClearCaches drops forge, origin-fetch, and local scan TTL state (for example after sync changes projects).
func (s *Service) ClearCaches() {
	if s == nil {
		return
	}
	if s.Cache != nil {
		s.Cache.Clear()
	}
	if s.OriginFetch != nil {
		s.OriginFetch.Clear()
	}
	s.scanMu.Lock()
	s.scanKey = ""
	s.scanAt = time.Time{}
	s.scanDisc = localgit.Discovery{}
	s.scanMu.Unlock()
}

// ToolingStatus returns forge CLI install/auth status for meta and Manage UI.
// Only GitHub and GitLab are probed; Azure DevOps and Bitbucket stay off the UI surface for now.
func (s *Service) ToolingStatus(ctx context.Context) board.Tooling {
	var out board.Tooling
	if s == nil {
		return out
	}
	auth := remotegit.NewAuthCache()
	if s.GitHub != nil {
		out.GitHub.Installed, out.GitHub.Authed, out.GitHub.Detail = auth.GetOrCheck("github", func() (bool, bool, string) {
			return s.GitHub.AuthStatus(ctx)
		})
	}
	if s.GitLab != nil {
		out.GitLab.Installed, out.GitLab.Authed, out.GitLab.Detail = auth.GetOrCheck("gitlab", func() (bool, bool, string) {
			return s.GitLab.AuthStatus(ctx)
		})
	}
	return out
}

// CollectOpts controls dashboard aggregation.
type CollectOpts struct {
	Fresh  bool
	ViewID string
}

// Collect builds the dashboard for projects in the resolved view.
// When fresh is true, forge and origin-fetch TTL caches are bypassed for this request.
// Empty viewID selects the first effective view. Unknown viewID returns an error.
func (s *Service) Collect(ctx context.Context, doc config.File, fresh bool, viewID string) (board.Dashboard, error) {
	return s.CollectWith(ctx, doc, CollectOpts{Fresh: fresh, ViewID: viewID})
}

// CollectWith builds the dashboard using CollectOpts.
func (s *Service) CollectWith(ctx context.Context, doc config.File, opts CollectOpts) (board.Dashboard, error) {
	return s.CollectStream(ctx, doc, opts, nil)
}

// CollectStream builds the dashboard like CollectWith. When onProject is set, it is
// called as each project row finishes (completion order, may be concurrent).
// The returned Dashboard.Projects slice is always in view membership order.
func (s *Service) CollectStream(ctx context.Context, doc config.File, opts CollectOpts, onProject func(board.ProjectSummary)) (board.Dashboard, error) {
	out := board.Dashboard{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Views:       ViewSummaries(doc),
	}
	projects, view, err := doc.ProjectsForView(opts.ViewID)
	if err != nil {
		return out, err
	}
	out.ActiveView = view.ID
	if s == nil {
		return out, nil
	}
	auth := remotegit.NewAuthCache()
	if s.GitHub != nil {
		installed, authed, detail := auth.GetOrCheck("github", func() (bool, bool, string) {
			return s.GitHub.AuthStatus(ctx)
		})
		out.Tooling.GitHub.Installed = installed
		out.Tooling.GitHub.Authed = authed
		out.Tooling.GitHub.Detail = detail
	}
	if s.GitLab != nil {
		installed, authed, detail := auth.GetOrCheck("gitlab", func() (bool, bool, string) {
			return s.GitLab.AuthStatus(ctx)
		})
		out.Tooling.GitLab.Installed = installed
		out.Tooling.GitLab.Authed = authed
		out.Tooling.GitLab.Detail = detail
	}
	if s.AzureDevOps != nil {
		installed, authed, detail := auth.GetOrCheck("azuredevops", func() (bool, bool, string) {
			return s.AzureDevOps.AuthStatus(ctx)
		})
		out.Tooling.AzureDevOps.Installed = installed
		out.Tooling.AzureDevOps.Authed = authed
		out.Tooling.AzureDevOps.Detail = detail
	}
	if s.Bitbucket != nil {
		installed, authed, detail := auth.GetOrCheck("bitbucket", func() (bool, bool, string) {
			return s.Bitbucket.AuthStatus(ctx)
		})
		out.Tooling.Bitbucket.Installed = installed
		out.Tooling.Bitbucket.Authed = authed
		out.Tooling.Bitbucket.Detail = detail
	}

	var disc localgit.Discovery
	if s.Local != nil && len(doc.Local.Roots) > 0 {
		scanTTL := time.Duration(doc.EffectiveScanSeconds()) * time.Second
		disc = s.scanRootsCached(ctx, doc.Local.Roots, scanTTL, opts.Fresh)
	}
	labelByKey := projectLabelsByKey(doc.Projects)

	summaryOpts := remotegit.SummaryOpts{
		Fresh:     opts.Fresh,
		Cache:     s.Cache,
		Auth:      auth,
		HeadsTTL:  time.Duration(doc.EffectiveHeadsSeconds()) * time.Second,
		MergedTTL: time.Duration(doc.EffectiveMergedSeconds()) * time.Second,
	}
	fetchTTL := time.Duration(doc.EffectiveFetchSeconds()) * time.Second

	rows := make([]board.ProjectSummary, len(projects))
	sem := make(chan struct{}, originFetchParallel)
	var wg sync.WaitGroup
	var emitMu sync.Mutex
	for i, p := range projects {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, p config.Project) {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				return
			}
			rowCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			row := s.summarize(rowCtx, p, summaryOpts)
			forgeBranches := filterHiddenBranches(row.Branches, doc.UI.HideBranches)
			row.Local = s.attachLocal(ctx, p, disc, labelByKey, opts.Fresh, fetchTTL, forgeBranchNames(forgeBranches))
			s.annotateContentOnDefault(ctx, &row)
			s.confirmMergedForCandidates(ctx, p, &row, summaryOpts)
			remotegit.EnrichPruneHints(&row)
			row.Branches = forgeBranches
			syncOpenItemsToVisibleBranches(&row)
			rows[i] = row
			if onProject != nil {
				emitMu.Lock()
				onProject(row)
				emitMu.Unlock()
			}
		}(i, p)
	}
	wg.Wait()
	out.Projects = rows
	return out, nil
}

func rootsCacheKey(roots []string) string {
	return strings.Join(roots, "\x00")
}

// scanRootsCached returns a TTL-gated Discovery for local roots.
// fresh bypasses the cache; a successful scan always replaces the entry.
func (s *Service) scanRootsCached(ctx context.Context, roots []string, ttl time.Duration, fresh bool) localgit.Discovery {
	if s == nil || s.Local == nil || len(roots) == 0 {
		return localgit.Discovery{}
	}
	key := rootsCacheKey(roots)
	if !fresh && ttl > 0 {
		s.scanMu.Lock()
		hit := s.scanKey == key && !s.scanAt.IsZero() && time.Since(s.scanAt) < ttl
		disc := s.scanDisc
		s.scanMu.Unlock()
		if hit {
			return disc
		}
	}
	disc := s.Local.ScanRoots(ctx, roots)
	if ctx.Err() != nil {
		return disc
	}
	s.scanMu.Lock()
	s.scanKey = key
	s.scanAt = time.Now()
	s.scanDisc = disc
	s.scanMu.Unlock()
	return disc
}

func projectLabelsByKey(projects []config.Project) map[string]string {
	out := make(map[string]string, len(projects))
	for _, p := range projects {
		key := localgit.TrackKey(string(p.Host), p.Path)
		label := strings.TrimSpace(p.Label)
		if label == "" {
			label = strings.TrimSpace(p.ID)
		}
		if label == "" {
			continue
		}
		out[key] = label
	}
	return out
}

func (s *Service) attachLocal(ctx context.Context, p config.Project, disc localgit.Discovery, labelByKey map[string]string, fresh bool, fetchTTL time.Duration, forgeBranches []string) *board.LocalStatus {
	if s.Local == nil {
		return nil
	}
	checkouts := disc.Appearances(string(p.Host), p.Path)
	explicit := strings.TrimSpace(p.LocalPath)
	if len(checkouts) == 0 && explicit == "" {
		if len(disc.ByKey) == 0 {
			return nil
		}
		return &board.LocalStatus{Mapped: false}
	}

	primary, ok := localgit.PickPrimary(explicit, checkouts)
	if !ok {
		return &board.LocalStatus{Mapped: false}
	}

	// Ensure primary path is in the inspect list even when local_path is outside scan.
	inspectList := checkouts
	if explicit != "" {
		absPrimary := primary.Path
		if expanded, err := localgit.ExpandPath(primary.Path); err == nil {
			absPrimary = expanded
			primary.Path = absPrimary
		}
		found := false
		for _, c := range inspectList {
			if filepath.Clean(c.Path) == filepath.Clean(absPrimary) {
				found = true
				break
			}
		}
		if !found {
			localgit.FillCheckoutMeta(&primary)
			inspectList = append([]localgit.Checkout{primary}, inspectList...)
		}
	}

	targetBranches := s.originFetchBranches(ctx, inspectList, forgeBranches)
	fetchErrByCommon := s.refreshOrigins(ctx, inspectList, fresh, fetchTTL, targetBranches)
	// Capture epochs after our own fetch bump so only concurrent mutations invalidate.
	epochsBeforeInspect := s.mutationEpochs(ctx, inspectList)

	type inspected struct {
		checkout localgit.Checkout
		status   localgit.Status
	}
	results := make([]inspected, len(inspectList))
	var wg sync.WaitGroup
	for i, c := range inspectList {
		wg.Add(1)
		go func(i int, c localgit.Checkout) {
			defer wg.Done()
			st := s.Local.InspectPath(ctx, c.Path)
			s.Local.EnrichOriginSync(ctx, &st)
			if st.Error == "" {
				if err := fetchErrFor(c, fetchErrByCommon, s.Local, ctx); err != nil {
					st.Error = err.Error()
					localgit.InvalidateOriginSync(&st)
				}
			}
			if s.mutationEpochChanged(ctx, c, epochsBeforeInspect) {
				localgit.InvalidateOriginSync(&st)
			}
			results[i] = inspected{checkout: c, status: st}
		}(i, c)
	}
	wg.Wait()

	parentLabelCache := map[string]string{}
	appearances := make([]board.LocalAppearance, 0, len(results))
	var union []board.LocalWorktree
	var primaryLocal *board.LocalStatus

	primaryPath := filepath.Clean(primary.Path)
	for _, r := range results {
		c := r.checkout
		parentLabel := ""
		if c.Role == localgit.RoleSubmodule && c.Superproject != "" {
			parentLabel = s.parentLabel(ctx, c.Superproject, labelByKey, parentLabelCache)
		}
		displayID := localgit.DisplayID(c, parentLabel)
		app := statusToAppearance(c, r.status, displayID, parentLabel)
		isPrimary := filepath.Clean(c.Path) == primaryPath
		app.Primary = isPrimary
		appearances = append(appearances, app)

		for _, wt := range app.Worktrees {
			if wt.Bare {
				continue
			}
			wt.AppearancePath = c.Path
			wt.AppearanceLabel = displayID
			union = append(union, wt)
		}

		if isPrimary {
			primaryLocal = toBoardLocal(r.status)
		}
	}

	if primaryLocal == nil {
		// Explicit path inspect may have failed matching; inspect primary alone.
		st := s.Local.InspectPath(ctx, primary.Path)
		s.Local.EnrichOriginSync(ctx, &st)
		if st.Error == "" {
			if err := fetchErrFor(primary, fetchErrByCommon, s.Local, ctx); err != nil {
				st.Error = err.Error()
				localgit.InvalidateOriginSync(&st)
			}
		}
		primaryLocal = toBoardLocal(st)
	}
	primaryLocal.Appearances = appearances
	primaryLocal.Worktrees = union
	return primaryLocal
}

// refreshOrigins runs TTL-gated origin refresh once per unique common git dir.
// Fresh or expired full TTL uses `git fetch --prune origin`. Warm TTL uses a
// targeted fetch of the supplied branches (empty list skips network).
// At most originFetchParallel fetches run at once to bound network load.
const originFetchParallel = 8

func (s *Service) refreshOrigins(ctx context.Context, checkouts []localgit.Checkout, fresh bool, fetchTTL time.Duration, branches []string) map[string]error {
	out := map[string]error{}
	if s.Local == nil || len(checkouts) == 0 {
		return out
	}
	type job struct {
		path   string
		common string
	}
	jobs := make([]job, 0, len(checkouts))
	seen := map[string]struct{}{}
	for _, c := range checkouts {
		path := c.Path
		if path == "" {
			continue
		}
		common := strings.TrimSpace(c.CommonGitDir)
		if common == "" {
			cd, err := localgit.ResolveCommonDir(ctx, s.Local.CommonGitDir, path)
			if err != nil {
				muCommon := filepath.Clean(path)
				out[muCommon] = err
				continue
			}
			common = cd
		}
		common = filepath.Clean(common)
		if _, ok := seen[common]; ok {
			continue
		}
		seen[common] = struct{}{}
		jobs = append(jobs, job{path: path, common: common})
	}

	limit := originFetchParallel
	if limit < 1 {
		limit = 1
	}
	sem := make(chan struct{}, limit)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := s.fetchOriginLocked(ctx, j.path, j.common, fetchTTL, fresh, branches); err != nil {
				mu.Lock()
				out[j.common] = err
				mu.Unlock()
			}
		}(j)
	}
	wg.Wait()
	return out
}

func (s *Service) fetchOriginLocked(ctx context.Context, path, common string, fetchTTL time.Duration, fresh bool, branches []string) error {
	lease, err := s.Mutations.Acquire(ctx, common)
	if err != nil {
		return err
	}
	defer lease.Release()
	updated, err := s.Local.FetchOriginSmart(ctx, path, fetchTTL, fresh, s.OriginFetch, branches)
	if err != nil {
		return err
	}
	if updated {
		lease.BumpEpoch()
	}
	return nil
}

func (s *Service) mutationEpochs(ctx context.Context, checkouts []localgit.Checkout) map[string]uint64 {
	out := map[string]uint64{}
	if s == nil || s.Mutations == nil {
		return out
	}
	for _, c := range checkouts {
		common := s.checkoutCommonDir(ctx, c)
		if common == "" {
			continue
		}
		if _, ok := out[common]; ok {
			continue
		}
		out[common] = s.Mutations.Epoch(common)
	}
	return out
}

func (s *Service) mutationEpochChanged(ctx context.Context, c localgit.Checkout, before map[string]uint64) bool {
	if s == nil || s.Mutations == nil || before == nil {
		return false
	}
	common := s.checkoutCommonDir(ctx, c)
	if common == "" {
		return false
	}
	prev, ok := before[common]
	if !ok {
		return false
	}
	return s.Mutations.Epoch(common) != prev
}

func (s *Service) checkoutCommonDir(ctx context.Context, c localgit.Checkout) string {
	common := strings.TrimSpace(c.CommonGitDir)
	if common != "" {
		return filepath.Clean(common)
	}
	if s.Local == nil || c.Path == "" {
		return ""
	}
	cd, err := localgit.ResolveCommonDir(ctx, s.Local.CommonGitDir, c.Path)
	if err != nil {
		return ""
	}
	return cd
}

func forgeBranchNames(branches []board.BranchRef) []string {
	out := make([]string, 0, len(branches))
	for _, b := range branches {
		name := strings.TrimSpace(b.Name)
		if name == "" {
			continue
		}
		out = append(out, name)
	}
	return out
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
	if s == nil || s.Local == nil {
		return out
	}
	for _, c := range checkouts {
		if c.Path == "" || ctx.Err() != nil {
			continue
		}
		heads, err := s.Local.ListLocalHeads(ctx, c.Path)
		if err != nil {
			continue
		}
		for _, name := range heads {
			add(name)
		}
	}
	return out
}

func fetchErrFor(c localgit.Checkout, byCommon map[string]error, in LocalGit, ctx context.Context) error {
	if len(byCommon) == 0 {
		return nil
	}
	common := strings.TrimSpace(c.CommonGitDir)
	if common == "" && in != nil && c.Path != "" {
		if cd, err := in.CommonGitDir(ctx, c.Path); err == nil && cd != "" {
			common = cd
		} else {
			common = c.Path
		}
	}
	common = filepath.Clean(common)
	if err, ok := byCommon[common]; ok {
		return err
	}
	return nil
}

func (s *Service) parentLabel(ctx context.Context, parentPath string, labelByKey map[string]string, cache map[string]string) string {
	parentPath = filepath.Clean(parentPath)
	if v, ok := cache[parentPath]; ok {
		return v
	}
	label := filepath.Base(parentPath)
	if s.Local != nil {
		url, err := s.Local.OriginRemote(ctx, parentPath)
		if err == nil {
			if ref, ok := localgit.ParseRemoteURL(url); ok {
				if lb, ok := labelByKey[localgit.TrackKey(ref.Host, ref.Path)]; ok {
					label = lb
				}
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
	for _, wt := range st.Worktrees {
		if wt.Bare {
			continue
		}
		app.Worktrees = append(app.Worktrees, board.LocalWorktree{
			Path:            wt.Path,
			Branch:          wt.Branch,
			Tag:             wt.Tag,
			Detached:        wt.Detached,
			Bare:            wt.Bare,
			Main:            wt.Main,
			Dirty:           wt.Dirty,
			Ahead:           wt.Ahead,
			Behind:          wt.Behind,
			Upstream:        wt.Upstream,
			AppearancePath:  c.Path,
			AppearanceLabel: displayID,
		})
	}
	return app
}

func toBoardLocal(st localgit.Status) *board.LocalStatus {
	out := &board.LocalStatus{
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
	}
	for _, wt := range st.Worktrees {
		if wt.Bare {
			continue
		}
		out.Worktrees = append(out.Worktrees, board.LocalWorktree{
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

func (s *Service) summarize(ctx context.Context, p config.Project, opts remotegit.SummaryOpts) board.ProjectSummary {
	missing := func(client string) board.ProjectSummary {
		return board.ProjectSummary{
			ID: p.ID, Label: p.Label, Host: string(p.Host), Path: p.Path, Org: forgeOrg(p.Path),
			OpenURL: p.OpenURL(), Capabilities: remotegit.HostCapabilities(p.Host), Error: client + " client missing",
		}
	}
	switch p.Host {
	case config.HostGitHub:
		if s.GitHub == nil {
			return missing("github")
		}
		row, _ := s.GitHub.ProjectSummary(ctx, p, opts)
		return row
	case config.HostGitLab:
		if s.GitLab == nil {
			return missing("gitlab")
		}
		row, _ := s.GitLab.ProjectSummary(ctx, p, opts)
		return row
	case config.HostAzureDevOps:
		if s.AzureDevOps == nil {
			return missing("azuredevops")
		}
		row, _ := s.AzureDevOps.ProjectSummary(ctx, p, opts)
		return row
	case config.HostBitbucket:
		if s.Bitbucket == nil {
			return missing("bitbucket")
		}
		row, _ := s.Bitbucket.ProjectSummary(ctx, p, opts)
		return row
	default:
		return board.ProjectSummary{
			ID: p.ID, Label: p.Label, Host: string(p.Host), Path: p.Path, Org: forgeOrg(p.Path),
			OpenURL: p.OpenURL(), Capabilities: remotegit.HostCapabilities(p.Host), Error: "unsupported host",
		}
	}
}

func forgeOrg(path string) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

// FindProject returns a project by id.
func FindProject(projects []config.Project, id string) (config.Project, bool) {
	for _, p := range projects {
		if p.ID == id {
			return p, true
		}
	}
	return config.Project{}, false
}

// ClientFor returns the forge client for a project host.
// A missing client is a true nil interface (not a typed nil pointer).
func ClientFor(s *Service, p config.Project) remotegit.Client {
	if s == nil {
		return nil
	}
	switch p.Host {
	case config.HostGitHub:
		if s.GitHub == nil {
			return nil
		}
		return s.GitHub
	case config.HostGitLab:
		if s.GitLab == nil {
			return nil
		}
		return s.GitLab
	case config.HostAzureDevOps:
		if s.AzureDevOps == nil {
			return nil
		}
		return s.AzureDevOps
	case config.HostBitbucket:
		if s.Bitbucket == nil {
			return nil
		}
		return s.Bitbucket
	default:
		return nil
	}
}
