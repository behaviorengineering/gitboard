// Package dashboard is public API: it aggregates tracked projects into board
// rows from forge clients (pkg/remotegit) and local checkouts (pkg/localgit).
// Service reads; Commands mutates local checkouts under a per-repository
// lease and edits the config document (pkg/boardconfig).
package dashboard

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/board"
	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/cliexec"
	"github.com/behaviorengineering/gitboard/pkg/clock"
	"github.com/behaviorengineering/gitboard/pkg/localgit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

// Service aggregates project rows via forge CLIs and optional local git.
// Build it with ServiceConfig.CreateService.
type Service struct {
	gitHub      *remotegit.GitHub
	gitLab      *remotegit.GitLab
	azureDevOps *remotegit.AzureDevOps
	bitbucket   *remotegit.Bitbucket
	local       LocalGit
	cache       *remotegit.TTLCache
	originFetch *localgit.OriginFetchCache
	mutations   *localgit.MutationCoordinator
	clock       clock.Clock

	scanMu   sync.Mutex
	scanKey  string
	scanAt   time.Time
	scanDisc localgit.Discovery
}

// ServiceConfig holds what a Service needs at construction time.
type ServiceConfig struct {
	// GitHub, GitLab, AzureDevOps and Bitbucket are the forge clients. Required.
	GitHub      *remotegit.GitHub
	GitLab      *remotegit.GitLab
	AzureDevOps *remotegit.AzureDevOps
	Bitbucket   *remotegit.Bitbucket
	// Local reads and mutates checkouts. Optional: nil for sync-only callers;
	// local columns and pull, prune and investigate commands are then off.
	Local LocalGit
	// Mutations serialises local git writes per common git directory.
	// Optional: nil builds one with the platform OS lock. Pass the same
	// coordinator to every service that writes the same repositories.
	Mutations *localgit.MutationCoordinator
	// Clock stamps responses and drives every TTL cache. Optional: nil uses
	// the wall clock.
	Clock clock.Clock
}

// DefaultServiceConfig returns a config with the four forge clients: GitHub,
// GitLab and Azure DevOps run their CLIs on run; Bitbucket calls its REST API
// with environment credentials. Callers set Local (for example
// NewLocalGit(run)) when they need local checkouts.
func DefaultServiceConfig(run cliexec.Exec) ServiceConfig {
	return ServiceConfig{
		GitHub:      remotegit.NewGitHub(run),
		GitLab:      remotegit.NewGitLab(run),
		AzureDevOps: remotegit.NewAzureDevOps(run),
		Bitbucket:   remotegit.DefaultBitbucketConfig().CreateBitbucket(),
	}
}

// CreateService validates the config and builds a Service with in-memory
// forge and origin-fetch caches on the config clock.
// Panics if a forge client is nil (required dependencies at wire-up).
func (cfg ServiceConfig) CreateService() *Service {
	if cfg.GitHub == nil || cfg.GitLab == nil || cfg.AzureDevOps == nil || cfg.Bitbucket == nil {
		panic("dashboard.CreateService: GitHub, GitLab, AzureDevOps and Bitbucket clients are required")
	}
	now := clock.OrWall(cfg.Clock)
	mutations := cfg.Mutations
	if mutations == nil {
		mutations = localgit.NewMutationCoordinator(localgit.DefaultOSLocker())
	}
	return &Service{
		gitHub:      cfg.GitHub,
		gitLab:      cfg.GitLab,
		azureDevOps: cfg.AzureDevOps,
		bitbucket:   cfg.Bitbucket,
		local:       cfg.Local,
		cache:       remotegit.NewTTLCache(now),
		originFetch: localgit.NewOriginFetchCache(now),
		mutations:   mutations,
		clock:       now,
	}
}

// ClearCaches drops forge, origin-fetch, and local scan TTL state (for example after sync changes projects).
func (s *Service) ClearCaches() {
	if s == nil {
		return
	}
	s.cache.Clear()
	s.originFetch.Clear()
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
	return s.toolingStatus(ctx, remotegit.NewAuthCache(), boardconfig.HostGitHub, boardconfig.HostGitLab)
}

// toolingStatus probes each host's forge client once through auth and fills
// its Tooling entry. Hosts not listed stay zero.
func (s *Service) toolingStatus(ctx context.Context, auth *remotegit.AuthCache, hosts ...boardconfig.Host) board.Tooling {
	var out board.Tooling
	for _, host := range hosts {
		c := s.clientFor(host)
		if c == nil {
			continue
		}
		installed, authed, detail := auth.GetOrCheck(host, func() (bool, bool, string) {
			return c.AuthStatus(ctx)
		})
		setToolStatus(&out, host, installed, authed, detail)
	}
	return out
}

// setToolStatus writes one forge's probe result into its Tooling field.
func setToolStatus(t *board.Tooling, host boardconfig.Host, installed, authed bool, detail string) {
	switch host {
	case boardconfig.HostGitHub:
		t.GitHub.Installed, t.GitHub.Authed, t.GitHub.Detail = installed, authed, detail
	case boardconfig.HostGitLab:
		t.GitLab.Installed, t.GitLab.Authed, t.GitLab.Detail = installed, authed, detail
	case boardconfig.HostAzureDevOps:
		t.AzureDevOps.Installed, t.AzureDevOps.Authed, t.AzureDevOps.Detail = installed, authed, detail
	case boardconfig.HostBitbucket:
		t.Bitbucket.Installed, t.Bitbucket.Authed, t.Bitbucket.Detail = installed, authed, detail
	}
}

// CollectOpts controls dashboard aggregation.
type CollectOpts struct {
	Fresh  bool
	ViewID string
	// RowBudget bounds each project's forge summary, so one slow forge
	// cannot hold the whole board. Zero leaves the caller's deadline as the
	// only bound. The HTTP handler sets it; library callers choose their own.
	RowBudget time.Duration
}

// Collect builds the dashboard for projects in the resolved view.
// When fresh is true, forge and origin-fetch TTL caches are bypassed for this request.
// Empty viewID selects the first effective view. Unknown viewID returns an error.
func (s *Service) Collect(ctx context.Context, doc boardconfig.File, fresh bool, viewID string) (board.Dashboard, error) {
	return s.CollectStream(ctx, doc, CollectOpts{Fresh: fresh, ViewID: viewID}, nil)
}

// CollectStream builds the dashboard like Collect. When onProject is set, it is
// called as each project row finishes (completion order, may be concurrent).
// The returned Dashboard.Projects slice is always in view membership order.
func (s *Service) CollectStream(ctx context.Context, doc boardconfig.File, opts CollectOpts, onProject func(board.ProjectSummary)) (board.Dashboard, error) {
	if s == nil {
		return board.Dashboard{}, unavailable("dashboard.CollectStream", ErrServiceMissing)
	}
	out := board.Dashboard{
		GeneratedAt: s.clock().UTC().Format(time.RFC3339),
		Views:       ViewSummaries(doc),
	}
	projects, view, err := doc.ProjectsForView(opts.ViewID)
	if err != nil {
		return out, invalidCause("dashboard.CollectStream", "unknown view", err)
	}
	out.ActiveView = view.ID
	auth := remotegit.NewAuthCache()
	out.Tooling = s.toolingStatus(ctx, auth, boardconfig.Hosts()...)

	var disc localgit.Discovery
	if s.local != nil && len(doc.Local.Roots) > 0 {
		scanTTL := time.Duration(doc.EffectiveScanSeconds()) * time.Second
		disc = s.scanRootsCached(ctx, doc.Local.Roots, scanTTL, opts.Fresh)
	}
	labelByKey := projectLabelsByKey(doc.Projects)

	summaryOpts := remotegit.SummaryOpts{
		Fresh:     opts.Fresh,
		Now:       s.clock,
		Cache:     s.cache,
		Auth:      auth,
		HeadsTTL:  time.Duration(doc.EffectiveHeadsSeconds()) * time.Second,
		MergedTTL: time.Duration(doc.EffectiveMergedSeconds()) * time.Second,
	}
	local := localInputs{
		disc:     disc,
		labels:   labelByKey,
		fresh:    opts.Fresh,
		fetchTTL: time.Duration(doc.EffectiveFetchSeconds()) * time.Second,
	}

	out.LocalWarnings = disc.Warnings
	row := rowInputs{summary: summaryOpts, local: local, budget: opts.RowBudget}

	rows := make([]board.ProjectSummary, len(projects))
	var emitMu sync.Mutex
	parallel(len(projects), collectParallel, func(i int) {
		if ctx.Err() != nil {
			return
		}
		rows[i] = s.collectRow(ctx, doc, projects[i], row)
		if onProject != nil {
			emitMu.Lock()
			onProject(rows[i])
			emitMu.Unlock()
		}
	})
	out.Projects = rows
	return out, nil
}

// rowInputs is what every project row of one collect shares.
type rowInputs struct {
	summary remotegit.SummaryOpts
	local   localInputs
	// budget bounds the forge summary; zero means the caller's deadline only.
	budget time.Duration
}

// collectRow builds one project's board row: the forge summary, the local
// checkouts, prune hints, then the hide_branches filter.
func (s *Service) collectRow(ctx context.Context, doc boardconfig.File, p boardconfig.Project, in rowInputs) board.ProjectSummary {
	summaryCtx := ctx
	if in.budget > 0 {
		var cancel context.CancelFunc
		summaryCtx, cancel = context.WithTimeout(ctx, in.budget)
		defer cancel()
	}
	row := s.summarize(summaryCtx, p, in.summary)
	forgeBranches := filterHiddenBranches(row.Branches, doc.UI.HideBranches)
	local := in.local
	local.forgeBranches = forgeBranchNames(forgeBranches)
	row.Local = s.attachLocal(ctx, p, local)
	s.annotatePrune(ctx, p, &row, in.summary)
	row.Branches = forgeBranches
	syncOpenItemsToVisibleBranches(&row)
	return row
}

// annotatePrune gathers prune evidence for the row's candidates (content on
// default, per-branch merged lookups) and sets every worktree's prune hint.
// It runs on the unfiltered forge branches, so a hidden branch row still
// blocks a prune.
func (s *Service) annotatePrune(ctx context.Context, p boardconfig.Project, row *board.ProjectSummary, opts remotegit.SummaryOpts) {
	s.annotateContentOnDefault(ctx, row)
	s.confirmMergedForCandidates(ctx, p, row, opts)
	enrichPruneHints(row)
}

func rootsCacheKey(roots []string) string {
	return strings.Join(roots, "\x00")
}

// scanRootsCached returns a TTL-gated Discovery for local roots.
// fresh bypasses the cache; a successful scan always replaces the entry.
func (s *Service) scanRootsCached(ctx context.Context, roots []string, ttl time.Duration, fresh bool) localgit.Discovery {
	if s == nil || s.local == nil || len(roots) == 0 {
		return localgit.Discovery{}
	}
	key := rootsCacheKey(roots)
	if !fresh && ttl > 0 {
		s.scanMu.Lock()
		hit := s.scanKey == key && !s.scanAt.IsZero() && s.clock().Sub(s.scanAt) < ttl
		disc := s.scanDisc
		s.scanMu.Unlock()
		if hit {
			return disc
		}
	}
	disc := s.local.ScanRoots(ctx, roots)
	if ctx.Err() != nil {
		return disc
	}
	s.scanMu.Lock()
	s.scanKey = key
	s.scanAt = s.clock()
	s.scanDisc = disc
	s.scanMu.Unlock()
	return disc
}

func projectLabelsByKey(projects []boardconfig.Project) map[string]string {
	out := make(map[string]string, len(projects))
	for _, p := range projects {
		key := localgit.TrackKey(p.Host, p.Path)
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

func (s *Service) summarize(ctx context.Context, p boardconfig.Project, opts remotegit.SummaryOpts) board.ProjectSummary {
	client := s.clientFor(p.Host)
	if client == nil {
		row := remotegit.BaseSummary(p)
		row.Error = "unsupported host"
		return row
	}
	// The row carries any load error in row.Error; the returned error only
	// repeats it for callers that need the chain.
	row, _ := client.ProjectSummary(ctx, p, opts)
	return row
}

// FindProject returns a project by id.
func FindProject(projects []boardconfig.Project, id string) (boardconfig.Project, bool) {
	for _, p := range projects {
		if p.ID == id {
			return p, true
		}
	}
	return boardconfig.Project{}, false
}

// ClientFor returns the forge client for a project host, or nil for an
// unsupported host or a nil Service.
func ClientFor(s *Service, p boardconfig.Project) remotegit.Client {
	if s == nil {
		return nil
	}
	return s.clientFor(p.Host)
}

func (s *Service) clientFor(host boardconfig.Host) remotegit.Client {
	switch host {
	case boardconfig.HostGitHub:
		return s.gitHub
	case boardconfig.HostGitLab:
		return s.gitLab
	case boardconfig.HostAzureDevOps:
		return s.azureDevOps
	case boardconfig.HostBitbucket:
		return s.bitbucket
	default:
		return nil
	}
}
