package dashboard

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/forge"
	"github.com/behaviorengineering/gitboard/internal/localgit"
)

// Service aggregates project rows via forge CLIs and optional local git.
type Service struct {
	GitHub      *forge.GitHub
	GitLab      *forge.GitLab
	Local       *localgit.Inspector
	Cache       *forge.TTLCache
	OriginFetch *localgit.OriginFetchCache
}

// New returns a dashboard service with in-memory upstream and origin-fetch caches.
func New(gh *forge.GitHub, gl *forge.GitLab, local *localgit.Inspector) *Service {
	return &Service{
		GitHub:      gh,
		GitLab:      gl,
		Local:       local,
		Cache:       forge.NewTTLCache(),
		OriginFetch: localgit.NewOriginFetchCache(),
	}
}

// ClearCaches drops forge and origin-fetch TTL state (for example after sync changes projects).
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
}

// Collect builds the dashboard for all configured projects.
// When fresh is true, forge and origin-fetch TTL caches are bypassed for this request.
func (s *Service) Collect(ctx context.Context, doc config.File, fresh bool) forge.Dashboard {
	projects := doc.Projects
	out := forge.Dashboard{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if s.GitHub != nil {
		installed, authed, detail := s.GitHub.AuthStatus(ctx)
		out.Tooling.GitHub.Installed = installed
		out.Tooling.GitHub.Authed = authed
		out.Tooling.GitHub.Detail = detail
	}
	if s.GitLab != nil {
		installed, authed, detail := s.GitLab.AuthStatus(ctx)
		out.Tooling.GitLab.Installed = installed
		out.Tooling.GitLab.Authed = authed
		out.Tooling.GitLab.Detail = detail
	}

	var disc localgit.Discovery
	if s.Local != nil && len(doc.Local.Roots) > 0 {
		disc = s.Local.ScanRoots(ctx, doc.Local.Roots)
	}
	labelByKey := projectLabelsByKey(doc.Projects)

	opts := forge.SummaryOpts{
		Fresh:     fresh,
		Cache:     s.Cache,
		HeadsTTL:  time.Duration(doc.EffectiveHeadsSeconds()) * time.Second,
		MergedTTL: time.Duration(doc.EffectiveMergedSeconds()) * time.Second,
	}
	fetchTTL := time.Duration(doc.EffectiveFetchSeconds()) * time.Second

	rows := make([]forge.ProjectSummary, len(projects))
	var wg sync.WaitGroup
	for i, p := range projects {
		wg.Add(1)
		go func(i int, p config.Project) {
			defer wg.Done()
			row := s.summarize(ctx, p, opts)
			row.Local = s.attachLocal(ctx, p, disc, labelByKey, fresh, fetchTTL)
			forge.EnrichPruneHints(&row)
			rows[i] = row
		}(i, p)
	}
	wg.Wait()
	out.Projects = rows
	return out
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

func (s *Service) attachLocal(ctx context.Context, p config.Project, disc localgit.Discovery, labelByKey map[string]string, fresh bool, fetchTTL time.Duration) *forge.LocalStatus {
	if s.Local == nil {
		return nil
	}
	checkouts := disc.Appearances(string(p.Host), p.Path)
	explicit := strings.TrimSpace(p.LocalPath)
	if len(checkouts) == 0 && explicit == "" {
		if len(disc.ByKey) == 0 {
			return nil
		}
		return &forge.LocalStatus{Mapped: false}
	}

	primary, ok := localgit.PickPrimary(explicit, checkouts)
	if !ok {
		return &forge.LocalStatus{Mapped: false}
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

	fetchErrByCommon := s.refreshOrigins(ctx, inspectList, fresh, fetchTTL)

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
			results[i] = inspected{checkout: c, status: st}
		}(i, c)
	}
	wg.Wait()

	parentLabelCache := map[string]string{}
	appearances := make([]forge.LocalAppearance, 0, len(results))
	var union []forge.LocalWorktree
	var primaryLocal *forge.LocalStatus

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
			primaryLocal = toForgeLocal(r.status)
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
		primaryLocal = toForgeLocal(st)
	}
	primaryLocal.Appearances = appearances
	primaryLocal.Worktrees = union
	return primaryLocal
}

// refreshOrigins runs TTL-gated git fetch origin once per unique common git dir.
// At most originFetchParallel fetches run at once to bound network load.
const originFetchParallel = 8

func (s *Service) refreshOrigins(ctx context.Context, checkouts []localgit.Checkout, fresh bool, fetchTTL time.Duration) map[string]error {
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
			if cd, err := s.Local.CommonGitDir(ctx, path); err == nil && cd != "" {
				common = cd
			} else {
				common = path
			}
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
			err := s.Local.FetchOriginCached(ctx, j.path, fetchTTL, fresh, s.OriginFetch)
			if err == nil {
				return
			}
			mu.Lock()
			out[j.common] = err
			mu.Unlock()
		}(j)
	}
	wg.Wait()
	return out
}

func fetchErrFor(c localgit.Checkout, byCommon map[string]error, in *localgit.Inspector, ctx context.Context) error {
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

func statusToAppearance(c localgit.Checkout, st localgit.Status, displayID, parentLabel string) forge.LocalAppearance {
	app := forge.LocalAppearance{
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
		OriginSync:    originSyncToForge(st.OriginSync),
	}
	if app.Role == "" {
		app.Role = forge.AppearanceStandalone
	}
	for _, wt := range st.Worktrees {
		if wt.Bare {
			continue
		}
		app.Worktrees = append(app.Worktrees, forge.LocalWorktree{
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

func toForgeLocal(st localgit.Status) *forge.LocalStatus {
	out := &forge.LocalStatus{
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
		OriginSync:    originSyncToForge(st.OriginSync),
	}
	for _, wt := range st.Worktrees {
		if wt.Bare {
			continue
		}
		out.Worktrees = append(out.Worktrees, forge.LocalWorktree{
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

func originSyncToForge(in []localgit.BranchSync) []forge.BranchOriginSync {
	if len(in) == 0 {
		return nil
	}
	out := make([]forge.BranchOriginSync, len(in))
	for i, s := range in {
		out[i] = forge.BranchOriginSync{Name: s.Name, Ahead: s.Ahead, Behind: s.Behind}
	}
	return out
}

func (s *Service) summarize(ctx context.Context, p config.Project, opts forge.SummaryOpts) forge.ProjectSummary {
	switch p.Host {
	case config.HostGitHub:
		if s.GitHub == nil {
			return forge.ProjectSummary{
				ID: p.ID, Label: p.Label, Host: string(p.Host), Path: p.Path, Org: forgeOrg(p.Path),
				OpenURL: p.OpenURL(), Error: "github client missing",
			}
		}
		row, _ := s.GitHub.ProjectSummary(ctx, p, opts)
		return row
	default:
		if s.GitLab == nil {
			return forge.ProjectSummary{
				ID: p.ID, Label: p.Label, Host: string(p.Host), Path: p.Path, Org: forgeOrg(p.Path),
				OpenURL: p.OpenURL(), Error: "gitlab client missing",
			}
		}
		row, _ := s.GitLab.ProjectSummary(ctx, p, opts)
		return row
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
func ClientFor(s *Service, p config.Project) forge.Client {
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
	default:
		return nil
	}
}
