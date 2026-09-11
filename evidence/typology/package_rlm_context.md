# Package RLM context index

AST-derived package context for recursive exploration.
Classify from symbols, imports, and bodies. Directory basename is not evidence.

## ./cmd/gitboard
- package: `main`
- packageDoc: Local code-change board for GitLab and GitHub (127.0.0.1 only).
- hasMain: true
- jsonTags: false
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: true
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- deliveryHint: cli
- mechanicalRole: entrypoint
- mechanicalConfidence: 0.90
- mechanicalEvidence: has_main
- exportedDecls: (none)
- exportedFuncs: (none)
- exportedMethods: (none)
- unexportedDecls: version
- unexportedFuncs: fileExists, main, printUsage, runInit, runServe, runSync
- unexportedMethods: (none)
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/cmd/gitboard/main.go

## ./internal/board
- package: `board`
- packageDoc: Package board contains the JSON data types shared across the dashboard UI, forge adapters, and server handlers.
- hasMain: false
- jsonTags: true
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- deliveryHint: dto
- mechanicalRole: dto
- mechanicalConfidence: 0.90
- mechanicalEvidence: json_tags, no_exported_funcs, no_exported_methods, shared_dto_multi_importer
- exportedDecls: AppearanceStandalone, AppearanceSubmodule, BranchOriginSync, BranchRef, CIStatus, Dashboard, FailedJob, LocalAppearance, LocalStatus, LocalWorktree, MergedReview, OpenItems, ProjectSummary, PruneLikely, PruneSafe, Tooling
- exportedFuncs: (none)
- exportedMethods: (none)
- unexportedDecls: (none)
- unexportedFuncs: (none)
- unexportedMethods: (none)
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/board/board.go

### Exported bodies

#### CIStatus (type)

```go
type CIStatus struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion,omitempty"`
	Ref        string `json:"ref,omitempty"`
	Name       string `json:"name,omitempty"`
	WebURL     string `json:"web_url,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	RunID      string `json:"run_id,omitempty"`
}
```

#### OpenItems (type)

```go
type OpenItems struct {
	PullRequests  int `json:"pull_requests"`
	MergeRequests int `json:"merge_requests"`
}
```

#### BranchRef (type)

```go
type BranchRef struct {
	Name       string `json:"name"`
	Default    bool   `json:"default,omitempty"`
	OpenReview bool   `json:"open_review,omitempty"`
	ReviewID   int    `json:"review_id,omitempty"`
	Draft      bool   `json:"draft,omitempty"`
	Conflict   bool   `json:"conflict,omitempty"`
	Stale      bool   `json:"stale,omitempty"`
	CIStatus   string `json:"ci_status,omitempty"`
	CIURL      string `json:"ci_url,omitempty"`
	RunID      string `json:"run_id,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	WebURL     string `json:"web_url,omitempty"`
}
```

#### MergedReview (type)

```go
type MergedReview struct {
	Branch   string `json:"branch"`
	ID       int    `json:"id,omitempty"`
	URL      string `json:"url,omitempty"`
	MergedAt string `json:"merged_at,omitempty"`
}
```

#### ProjectSummary (type)

```go
type ProjectSummary struct {
	ID        string         `json:"id"`
	Label     string         `json:"label"`
	Host      string         `json:"host"`
	Path      string         `json:"path,omitempty"`
	Org       string         `json:"org,omitempty"`
	OpenURL   string         `json:"open_url"`
	CI        *CIStatus      `json:"ci,omitempty"`
	Branches  []BranchRef    `json:"branches,omitempty"`
	Merged    []MergedReview `json:"merged,omitempty"`
	OpenItems OpenItems      `json:"open_items"`
	Local     *LocalStatus   `json:"local,omitempty"`
	Error     string         `json:"error,omitempty"`

	// RemoteNames is the full remote head set for prune matching (not sent to UI).
	// Loaded via forge API calls; may be incomplete only when RemoteNamesOK is false.
	RemoteNames []string `json:"-"`
	// RemoteNamesOK is true when heads were loaded successfully (live or cache).
	// When false, EnrichPruneHints suppresses all prune hints (fail closed).
	RemoteNamesOK bool `json:"-"`
	// MergedOK is true when the merged slice was loaded successfully (live or fresh cache).
	MergedOK bool `json:"-"`
}
```

#### LocalAppearance (type)

```go
type LocalAppearance struct {
	Role          string             `json:"role,omitempty"` // standalone | submodule
	Path          string             `json:"path,omitempty"`
	DisplayID     string             `json:"display_id,omitempty"`
	ParentPath    string             `json:"parent_path,omitempty"`
	ParentLabel   string             `json:"parent_label,omitempty"`
	RelPath       string             `json:"rel_path,omitempty"`
	Primary       bool               `json:"primary,omitempty"`
	Error         string             `json:"error,omitempty"`
	Branch        string             `json:"branch,omitempty"`
	Tag           string             `json:"tag,omitempty"` // Exact tag when Detached and HEAD is tagged.
	Detached      bool               `json:"detached,omitempty"`
	Dirty         bool               `json:"dirty,omitempty"`
	Ahead         int                `json:"ahead,omitempty"`
	Behind        int                `json:"behind,omitempty"`
	Upstream      string             `json:"upstream,omitempty"`
	DefaultBranch string             `json:"default_branch,omitempty"`
	DefaultBehind int                `json:"default_behind,omitempty"`
	DefaultAhead  int                `json:"default_ahead,omitempty"`
	OriginSync    []BranchOriginSync `json:"origin_sync,omitempty"`
	Worktrees     []LocalWorktree    `json:"worktrees,omitempty"`
}
```

#### LocalStatus (type)

```go
type LocalStatus struct {
	Mapped        bool               `json:"mapped"`
	Path          string             `json:"path,omitempty"`
	Error         string             `json:"error,omitempty"`
	Branch        string             `json:"branch,omitempty"`
	Tag           string             `json:"tag,omitempty"` // Exact tag when Detached and HEAD is tagged.
	Detached      bool               `json:"detached,omitempty"`
	Dirty         bool               `json:"dirty,omitempty"`
	Ahead         int                `json:"ahead,omitempty"`
	Behind        int                `json:"behind,omitempty"`
	Upstream      string             `json:"upstream,omitempty"`
	DefaultBranch string             `json:"default_branch,omitempty"`
	DefaultBehind int                `json:"default_behind,omitempty"`
	DefaultAhead  int                `json:"default_ahead,omitempty"`
	OriginSync    []BranchOriginSync `json:"origin_sync,omitempty"`
	Worktrees     []LocalWorktree    `json:"worktrees,omitempty"`
	Appearances   []LocalAppearance  `json:"appearances,omitempty"`
}
```

#### BranchOriginSync (type)

```go
type BranchOriginSync struct {
	Name   string `json:"name"`
	Ahead  int    `json:"ahead,omitempty"`
	Behind int    `json:"behind,omitempty"`
}
```

#### LocalWorktree (type)

```go
type LocalWorktree struct {
	Path             string `json:"path"`
	Branch           string `json:"branch,omitempty"`
	Tag              string `json:"tag,omitempty"` // Exact tag when Detached and HEAD is tagged.
	Detached         bool   `json:"detached,omitempty"`
	Bare             bool   `json:"bare,omitempty"`
	Main             bool   `json:"main,omitempty"`
	Dirty            bool   `json:"dirty,omitempty"`
	Ahead            int    `json:"ahead,omitempty"`
	Behind           int    `json:"behind,omitempty"`
	Upstream         string `json:"upstream,omitempty"`
	PruneHint        string `json:"prune_hint,omitempty"` // safe | likely
	ContentOnDefault bool   `json:"content_on_default,omitempty"`
	MergedID         int    `json:"merged_id,omitempty"`
	MergedURL        string `json:"merged_url,omitempty"`
	MergedAt         string `json:"merged_at,omitempty"`
	AppearancePath   string `json:"appearance_path,omitempty"`
	AppearanceLabel  string `json:"appearance_label,omitempty"`
}
```

#### FailedJob (type)

```go
type FailedJob struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Stage  string `json:"stage,omitempty"`
	WebURL string `json:"web_url,omitempty"`
}
```

#### Tooling (type)

```go
type Tooling struct {
	GitHub struct {
		Installed bool   `json:"installed"`
		Authed    bool   `json:"authed"`
		Detail    string `json:"detail,omitempty"`
	} `json:"github"`
	GitLab struct {
		Installed bool   `json:"installed"`
		Authed    bool   `json:"authed"`
		Detail    string `json:"detail,omitempty"`
	} `json:"gitlab"`
}
```

#### Dashboard (type)

```go
type Dashboard struct {
	GeneratedAt         string           `json:"generated_at"`
	PollIntervalSeconds int              `json:"poll_interval_seconds"`
	Tooling             Tooling          `json:"tooling"`
	Projects            []ProjectSummary `json:"projects"`
}
```


## ./internal/cliexec
- package: `cliexec`
- hasMain: false
- jsonTags: false
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: true
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- mechanicalRole: exec_runner
- mechanicalConfidence: 0.80
- mechanicalEvidence: imports_os_exec, exports_run_surface
- exportedDecls: Exec, Runner
- exportedFuncs: New
- exportedMethods: Runner.LookPath, Runner.Run, Runner.RunJSON
- unexportedDecls: (none)
- unexportedFuncs: (none)
- unexportedMethods: (none)
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/cliexec/cliexec.go

### Exported bodies

#### Exec (type)

```go
type Exec interface {
	LookPath(name string) (string, error)
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
	RunJSON(ctx context.Context, name string, args ...string) ([]byte, error)
}
```

#### Runner (type)

```go
type Runner struct {
	Timeout time.Duration
}
```

#### New (func)

```go
func New() *Runner {
	return &Runner{Timeout: 45 * time.Second}
}
```

#### Runner.LookPath (method)

```go
func (r *Runner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
```

#### Runner.RunJSON (method)

```go
func (r *Runner) RunJSON(ctx context.Context, name string, args ...string) ([]byte, error) {
	out, err := r.Run(ctx, name, args...)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return []byte("[]"), nil
	}
	return []byte(trimmed), nil
}
```

#### Runner.Run (method)

```go
func (r *Runner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if r == nil {
		r = New()
	}
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s %s: %s: %w", name, strings.Join(args, " "), msg, err)
	}
	return stdout.Bytes(), nil
}
```


## ./internal/config
- package: `config`
- hasMain: false
- jsonTags: true
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- mechanicalRole: config
- mechanicalConfidence: 0.80
- mechanicalEvidence: load_save_exports
- exportedDecls: DefaultExample, DefaultFetchSeconds, DefaultHeadsSeconds, DefaultLLMModel, DefaultMergedSeconds, DefaultPollSeconds, FailureDump, File, GitHubSync, GitLabSync, Host, HostGitHub, HostGitLab, LLM, Local, OpenInference, Project, SyncSources, UI, Upstream
- exportedFuncs: AgentsDir, DefaultPath, Dir, Init, Load, Save
- exportedMethods: File.EffectiveFetchSeconds, File.EffectiveHeadsSeconds, File.EffectiveLLM, File.EffectiveMergedSeconds, File.EffectiveOpenInference, File.EffectivePollSeconds, Project.OpenURL, SyncSources.HasSyncSources
- unexportedDecls: (none)
- unexportedFuncs: expandHome, fileExists, firstNonEmpty, normalizeLocal, normalizeSync, trimNonEmpty, validateProject
- unexportedMethods: (none)
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/config/config.go

### Exported bodies

#### Host (type)

```go
type Host string
```

#### Project (type)

```go
type Project struct {
	ID        string `yaml:"id" json:"id"`
	Label     string `yaml:"label" json:"label"`
	Host      Host   `yaml:"host" json:"host"`
	Path      string `yaml:"path" json:"path"`
	LocalPath string `yaml:"local_path,omitempty" json:"local_path,omitempty"`
}
```

#### Local (type)

```go
type Local struct {
	// Roots are directories to scan for git checkouts (and worktrees).
	// Matching uses origin remote URL against project host+path.
	Roots []string `yaml:"roots"`
	// FetchSeconds TTL-gates git fetch origin before local↔origin sync.
	// nil → default 120; explicit 0 always fetches; negative → default.
	FetchSeconds *int `yaml:"fetch_seconds"`
}
```

#### LLM (type)

```go
type LLM struct {
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
	APIKey  string `yaml:"api_key"`
}
```

#### FailureDump (type)

```go
type FailureDump struct {
	Enabled     *bool  `yaml:"enabled"`
	Dir         string `yaml:"dir"`
	MaxAgeHours int    `yaml:"max_age_hours"`
	MaxFiles    int    `yaml:"max_files"`
}
```

#### OpenInference (type)

```go
type OpenInference struct {
	Enabled     *bool       `yaml:"enabled"`
	Endpoint    string      `yaml:"endpoint"`
	ServiceName string      `yaml:"service_name"`
	FailureDump FailureDump `yaml:"failure_dump"`
}
```

#### GitHubSync (type)

```go
type GitHubSync struct {
	Orgs []string `yaml:"orgs"`
}
```

#### GitLabSync (type)

```go
type GitLabSync struct {
	Groups []string `yaml:"groups"`
}
```

#### SyncSources (type)

```go
type SyncSources struct {
	GitHub GitHubSync `yaml:"github"`
	GitLab GitLabSync `yaml:"gitlab"`
}
```

#### UI (type)

```go
type UI struct {
	// PollSeconds is the browser auto-refresh interval in seconds.
	// nil / omitted → default 30; 0 disables polling.
	PollSeconds *int `yaml:"poll_seconds"`
	// HideBranches are path.Match patterns for branch names omitted from the
	// dashboard Branches column. Empty / omitted → show all (default off).
	HideBranches []string `yaml:"hide_branches,omitempty" json:"hide_branches,omitempty"`
}
```

#### Upstream (type)

```go
type Upstream struct {
	// HeadsSeconds caches remote branch list + default branch. nil → 120.
	HeadsSeconds *int `yaml:"heads_seconds"`
	// MergedSeconds caches merged PR/MR lists for prune hints. nil → 600.
	MergedSeconds *int `yaml:"merged_seconds"`
}
```

#### File (type)

```go
type File struct {
	LLM           LLM           `yaml:"llm"`
	OpenInference OpenInference `yaml:"openinference"`
	UI            UI            `yaml:"ui"`
	Upstream      Upstream      `yaml:"upstream"`
	Local         Local         `yaml:"local"`
	Sync          SyncSources   `yaml:"sync"`
	Projects      []Project     `yaml:"projects"`
}
```

#### Dir (func)

```go
func Dir() string {
	if v := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); v != "" {
		return filepath.Join(v, "gitboard")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".config", "gitboard")
	}
	return filepath.Join(home, ".config", "gitboard")
}
```

#### DefaultPath (func)

```go
func DefaultPath() string {
	if v := strings.TrimSpace(os.Getenv("GITBOARD_CONFIG")); v != "" {
		return expandHome(v)
	}
	if v := strings.TrimSpace(os.Getenv("GITBOARD_PROJECTS")); v != "" {
		return expandHome(v)
	}
	userPath := filepath.Join(Dir(), "config.yaml")
	if fileExists(userPath) {
		return userPath
	}
	cwd, cwdErr := os.Getwd()
	if cwdErr == nil && cwd != "" {
		for _, name := range []string{"config.yaml", "projects.yaml"} {
			p := filepath.Join(cwd, name)
			if fileExists(p) {
				return p
			}
		}
	}
	return userPath
}
```

#### Load (func)

```go
func Load(path string) (File, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return File{}, fmt.Errorf("read config: %w", err)
	}
	var doc File
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return File{}, fmt.Errorf("parse config: %w", err)
	}
	for i, p := range doc.Projects {
		if err := validateProject(p); err != nil {
			return File{}, fmt.Errorf("projects[%d]: %w", i, err)
		}
		doc.Projects[i].Host = Host(strings.ToLower(string(p.Host)))
		doc.Projects[i].LocalPath = strings.TrimSpace(p.LocalPath)
	}
	normalizeSync(&doc.Sync)
	normalizeLocal(&doc.Local)
	return doc, nil
}
```

#### Save (func)

```go
func Save(path string, doc File) error {
	normalizeSync(&doc.Sync)
	normalizeLocal(&doc.Local)
	for i, p := range doc.Projects {
		if err := validateProject(p); err != nil {
			return fmt.Errorf("projects[%d]: %w", i, err)
		}
		doc.Projects[i].Host = Host(strings.ToLower(string(p.Host)))
		doc.Projects[i].LocalPath = strings.TrimSpace(p.LocalPath)
	}
	raw, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("mkdir config: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
```

#### Init (func)

```go
func Init(path string) (created bool, err error) {
	if fileExists(path) {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return false, fmt.Errorf("mkdir config: %w", err)
	}
	if err := os.WriteFile(path, []byte(DefaultExample), 0o600); err != nil {
		return false, fmt.Errorf("write config: %w", err)
	}
	return true, nil
}
```

#### SyncSources.HasSyncSources (method)

```go
func (s SyncSources) HasSyncSources() bool {
	return len(s.GitHub.Orgs) > 0 || len(s.GitLab.Groups) > 0
}
```

#### Project.OpenURL (method)

```go
func (p Project) OpenURL() string {
	path := strings.Trim(p.Path, "/")
	switch p.Host {
	case HostGitHub:
		return "https://github.com/" + path
	default:
		return "https://gitlab.com/" + path
	}
}
```

#### File.EffectiveLLM (method)

```go
func (f File) EffectiveLLM() LLM {
	base := firstNonEmpty(
		os.Getenv("GITBOARD_LLM_BASE_URL"),
		os.Getenv("POLYPUS_BASE_URL"),
		f.LLM.BaseURL,
	)
	if base != "" && !strings.HasSuffix(base, "/v1") {
		base = strings.TrimRight(base, "/") + "/v1"
	}
	return LLM{
		BaseURL: strings.TrimRight(base, "/"),
		APIKey: firstNonEmpty(
			os.Getenv("GITBOARD_LLM_API_KEY"),
			os.Getenv("OPENAI_API_KEY"),
			f.LLM.APIKey,
		),
		Model: firstNonEmpty(
			os.Getenv("GITBOARD_LLM_MODEL"),
			f.LLM.Model,
			DefaultLLMModel,
		),
	}
}
```

#### File.EffectivePollSeconds (method)

```go
func (f File) EffectivePollSeconds() int {
	if v := strings.TrimSpace(os.Getenv("GITBOARD_POLL_SECONDS")); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n >= 0 {
			return n
		}
	}
	if f.UI.PollSeconds == nil {
		return DefaultPollSeconds
	}
	if *f.UI.PollSeconds < 0 {
		return DefaultPollSeconds
	}
	return *f.UI.PollSeconds
}
```

#### File.EffectiveHeadsSeconds (method)

```go
func (f File) EffectiveHeadsSeconds() int {
	if f.Upstream.HeadsSeconds == nil {
		return DefaultHeadsSeconds
	}
	if *f.Upstream.HeadsSeconds < 0 {
		return DefaultHeadsSeconds
	}
	return *f.Upstream.HeadsSeconds
}
```

#### File.EffectiveMergedSeconds (method)

```go
func (f File) EffectiveMergedSeconds() int {
	if f.Upstream.MergedSeconds == nil {
		return DefaultMergedSeconds
	}
	if *f.Upstream.MergedSeconds < 0 {
		return DefaultMergedSeconds
	}
	return *f.Upstream.MergedSeconds
}
```

#### File.EffectiveFetchSeconds (method)

```go
func (f File) EffectiveFetchSeconds() int {
	if f.Local.FetchSeconds == nil {
		return DefaultFetchSeconds
	}
	if *f.Local.FetchSeconds < 0 {
		return DefaultFetchSeconds
	}
	return *f.Local.FetchSeconds
}
```

#### AgentsDir (func)

```go
func AgentsDir() string {
	return filepath.Join(Dir(), "agents")
}
```

#### File.EffectiveOpenInference (method)

```go
func (f File) EffectiveOpenInference() OpenInference {
	enabled := true
	if f.OpenInference.Enabled != nil {
		enabled = *f.OpenInference.Enabled
	}
	if v := strings.TrimSpace(os.Getenv("GITBOARD_OPENINFERENCE_ENABLED")); v != "" {
		enabled = v == "1" || strings.EqualFold(v, "true")
	}
	dumpEnabled := true
	if f.OpenInference.FailureDump.Enabled != nil {
		dumpEnabled = *f.OpenInference.FailureDump.Enabled
	}
	en := enabled
	den := dumpEnabled
	dir := strings.TrimSpace(f.OpenInference.FailureDump.Dir)
	if dir == "" {
		dir = filepath.Join(Dir(), "logs", "inference-failures")
	}
	maxAge := f.OpenInference.FailureDump.MaxAgeHours
	if maxAge <= 0 {
		maxAge = 48
	}
	maxFiles := f.OpenInference.FailureDump.MaxFiles
	if maxFiles <= 0 {
		maxFiles = 20
	}
	return OpenInference{
		Enabled:     &en,
		Endpoint:    firstNonEmpty(os.Getenv("GITBOARD_OPENINFERENCE_ENDPOINT"), f.OpenInference.Endpoint),
		ServiceName: firstNonEmpty(f.OpenInference.ServiceName, "gitboard"),
		FailureDump: FailureDump{
			Enabled:     &den,
			Dir:         dir,
			MaxAgeHours: maxAge,
			MaxFiles:    maxFiles,
		},
	}
}
```

### Private one-hop bodies

#### expandHome (func)

```go
func expandHome(p string) string {
	p = strings.TrimSpace(p)
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return p
		}
		if p == "~" {
			return home
		}
		return filepath.Join(home, p[2:])
	}
	return p
}
```

#### fileExists (func)

```go
func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
```

#### firstNonEmpty (func)

```go
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
```

#### normalizeLocal (func)

```go
func normalizeLocal(l *Local) {
	l.Roots = trimNonEmpty(l.Roots)
}
```

#### normalizeSync (func)

```go
func normalizeSync(s *SyncSources) {
	s.GitHub.Orgs = trimNonEmpty(s.GitHub.Orgs)
	s.GitLab.Groups = trimNonEmpty(s.GitLab.Groups)
}
```

#### validateProject (func)

```go
func validateProject(p Project) error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("missing id")
	}
	if strings.TrimSpace(p.Label) == "" {
		return fmt.Errorf("missing label")
	}
	switch Host(strings.ToLower(string(p.Host))) {
	case HostGitHub, HostGitLab:
	default:
		return fmt.Errorf("host must be github or gitlab")
	}
	parts := strings.Split(strings.Trim(p.Path, "/"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("path must be owner/repo")
	}
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("path must be owner/repo")
		}
	}
	return nil
}
```


## ./internal/dashboard
- package: `dashboard`
- hasMain: false
- jsonTags: true
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- mechanicalRole: aggregator
- mechanicalConfidence: 0.80
- mechanicalEvidence: orchestration_export, internal_imports_ge_2
- exportedDecls: BadRequestError, Commands, LocalGit, PruneSafeRequest, PullFFRequest, PullFFResult, Service, SyncInvestigation, SyncInvestigationRequest
- exportedFuncs: ClientFor, FindProject, IsBadRequest, New, NewCommands
- exportedMethods: BadRequestError.Error, Commands.InvestigateSync, Commands.PruneSafe, Commands.PullFF, Commands.RequireMappedPath, Service.ClearCaches, Service.Collect
- unexportedDecls: _, originFetchParallel
- unexportedFuncs: badRequest, branchHidden, fetchErrFor, filterHiddenBranches, findSafeWorktree, forgeOrg, originSyncToBoard, projectLabelsByKey, repoPathAllowed, statusToAppearance, toBoardLocal
- unexportedMethods: Service.annotateContentOnDefault, Service.attachLocal, Service.parentLabel, Service.refreshOrigins, Service.summarize
- errorTypes: BadRequestError
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/dashboard/commands.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/dashboard/content_prune.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/dashboard/dashboard.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/dashboard/hide_branches.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/dashboard/local.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/dashboard/prune.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/dashboard/pull.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/dashboard/sync.go

### Exported bodies

#### Commands (type)

```go
type Commands struct {
	*Service
}
```

#### NewCommands (func)

```go
func NewCommands(s *Service) *Commands {
	return &Commands{Service: s}
}
```

#### Service (type)

```go
type Service struct {
	GitHub      *remotegit.GitHub
	GitLab      *remotegit.GitLab
	Local       LocalGit
	Cache       *remotegit.TTLCache
	OriginFetch *localgit.OriginFetchCache
}
```

#### New (func)

```go
func New(gh *remotegit.GitHub, gl *remotegit.GitLab, local LocalGit) *Service {
	return &Service{
		GitHub:      gh,
		GitLab:      gl,
		Local:       local,
		Cache:       remotegit.NewTTLCache(),
		OriginFetch: localgit.NewOriginFetchCache(),
	}
}
```

#### Service.ClearCaches (method)

```go
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
```

#### Service.Collect (method)

```go
func (s *Service) Collect(ctx context.Context, doc config.File, fresh bool) board.Dashboard {
	out := board.Dashboard{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if s == nil {
		return out
	}
	projects := doc.Projects
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

	opts := remotegit.SummaryOpts{
		Fresh:     fresh,
		Cache:     s.Cache,
		HeadsTTL:  time.Duration(doc.EffectiveHeadsSeconds()) * time.Second,
		MergedTTL: time.Duration(doc.EffectiveMergedSeconds()) * time.Second,
	}
	fetchTTL := time.Duration(doc.EffectiveFetchSeconds()) * time.Second

	rows := make([]board.ProjectSummary, len(projects))
	sem := make(chan struct{}, originFetchParallel)
	var wg sync.WaitGroup
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
			row := s.summarize(ctx, p, opts)
			row.Local = s.attachLocal(ctx, p, disc, labelByKey, fresh, fetchTTL)
			s.annotateContentOnDefault(ctx, &row)
			remotegit.EnrichPruneHints(&row)
			row.Branches = filterHiddenBranches(row.Branches, doc.UI.HideBranches)
			rows[i] = row
		}(i, p)
	}
	wg.Wait()
	out.Projects = rows
	return out
}
```

#### FindProject (func)

```go
func FindProject(projects []config.Project, id string) (config.Project, bool) {
	for _, p := range projects {
		if p.ID == id {
			return p, true
		}
	}
	return config.Project{}, false
}
```

#### ClientFor (func)

```go
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
	default:
		return nil
	}
}
```

#### LocalGit (type)

```go
type LocalGit interface {
	ScanRoots(ctx context.Context, roots []string) localgit.Discovery
	InspectPath(ctx context.Context, path string) localgit.Status
	EnrichOriginSync(ctx context.Context, st *localgit.Status)
	CommonGitDir(ctx context.Context, repoPath string) (string, error)
	FetchOriginCached(ctx context.Context, repoPath string, ttl time.Duration, fresh bool, cache *localgit.OriginFetchCache) error
	OriginRemote(ctx context.Context, dir string) (string, error)
	PullFFOnly(ctx context.Context, repoPath, branch string) error
	RemoveSafeCheckout(ctx context.Context, worktreePath, branch, defaultBranch string) error
	ContentOnDefault(ctx context.Context, repoPath, branch, defaultBranch string) (ok bool, reason string, err error)
	InspectSync(ctx context.Context, repoPath, branch string) (localgit.SyncInspection, error)
}
```

#### BadRequestError (error)

```go
type BadRequestError struct {
	Msg string
}
```

#### BadRequestError.Error (method)

```go
func (e BadRequestError) Error() string {
	if e.Msg == "" {
		return "bad request"
	}
	return e.Msg
}
```

#### IsBadRequest (func)

```go
func IsBadRequest(err error) bool {
	var e BadRequestError
	return errors.As(err, &e)
}
```

#### PruneSafeRequest (type)

```go
type PruneSafeRequest struct {
	ProjectID    string `json:"project_id"`
	Branch       string `json:"branch"`
	WorktreePath string `json:"worktree_path"`
}
```

#### Commands.PruneSafe (method)

```go
func (c *Commands) PruneSafe(ctx context.Context, doc config.File, req PruneSafeRequest) error {
	s := c.Service
	if s == nil || s.Local == nil {
		return fmt.Errorf("local git inspector missing")
	}
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	worktreePath := strings.TrimSpace(req.WorktreePath)
	if projectID == "" || branch == "" || worktreePath == "" {
		return badRequest("project_id, branch, and worktree_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return badRequest(err.Error())
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return badRequest("unknown project")
	}

	abs, err := localgit.ExpandPath(worktreePath)
	if err != nil {
		return badRequest(fmt.Sprintf("worktree path: %v", err))
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	opts := remotegit.SummaryOpts{
		Fresh:     true,
		Cache:     s.Cache,
		HeadsTTL:  0,
		MergedTTL: 0,
	}
	row := s.summarize(ctx, p, opts)
	if !row.RemoteNamesOK {
		return badRequest("cannot re-validate remote heads; refuse prune")
	}
	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		disc = s.Local.ScanRoots(ctx, doc.Local.Roots)
	}
	row.Local = s.attachLocal(ctx, p, disc, projectLabelsByKey(doc.Projects), true, time.Duration(doc.EffectiveFetchSeconds())*time.Second)
	if row.Local == nil || !row.Local.Mapped {
		return badRequest("project has no mapped local checkout")
	}
	if !repoPathAllowed(row.Local, abs) {
		return badRequest("worktree_path is not a mapped checkout for this project")
	}
	s.annotateContentOnDefault(ctx, &row)
	remotegit.EnrichPruneHints(&row)

	wt, ok := findSafeWorktree(row.Local, branch, abs)
	if !ok {
		return badRequest(fmt.Sprintf("branch %q is not safe to remove (re-check prune hints)", branch))
	}
	defaultBranch := strings.TrimSpace(row.Local.DefaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	if err := s.Local.RemoveSafeCheckout(ctx, wt.Path, branch, defaultBranch); err != nil {
		if errors.Is(err, localgit.ErrInvalidBranch) ||
			errors.Is(err, localgit.ErrDirtyTree) ||
			errors.Is(err, localgit.ErrDiverged) ||
			errors.Is(err, localgit.ErrMissingBranch) {
			return badRequest(err.Error())
		}
		return fmt.Errorf("remove checkout: %w", err)
	}
	return nil
}
```

#### PullFFRequest (type)

```go
type PullFFRequest struct {
	ProjectID string `json:"project_id"`
	Branch    string `json:"branch"`
	RepoPath  string `json:"repo_path"`
}
```

#### PullFFResult (type)

```go
type PullFFResult struct {
	OK     bool   `json:"ok"`
	Branch string `json:"branch"`
	Path   string `json:"path"`
}
```

#### Commands.PullFF (method)

```go
func (c *Commands) PullFF(ctx context.Context, doc config.File, req PullFFRequest) (PullFFResult, error) {
	var zero PullFFResult
	s := c.Service
	if s == nil || s.Local == nil {
		return zero, fmt.Errorf("local git inspector missing")
	}
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	repoPath := strings.TrimSpace(req.RepoPath)
	if projectID == "" || branch == "" || repoPath == "" {
		return zero, badRequest("project_id, branch, and repo_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return zero, badRequest(err.Error())
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return zero, badRequest("unknown project")
	}

	abs, err := localgit.ExpandPath(repoPath)
	if err != nil {
		return zero, badRequest(fmt.Sprintf("repo path: %v", err))
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		disc = s.Local.ScanRoots(ctx, doc.Local.Roots)
	}
	local := s.attachLocal(ctx, p, disc, projectLabelsByKey(doc.Projects), false, time.Duration(doc.EffectiveFetchSeconds())*time.Second)
	if local == nil || !local.Mapped {
		return zero, badRequest("project has no mapped local checkout")
	}
	if !repoPathAllowed(local, abs) {
		return zero, badRequest("repo_path is not a mapped checkout for this project")
	}

	if err := s.Local.PullFFOnly(ctx, abs, branch); err != nil {
		if errors.Is(err, localgit.ErrInvalidBranch) ||
			errors.Is(err, localgit.ErrMissingBranch) ||
			errors.Is(err, localgit.ErrDirtyTree) ||
			errors.Is(err, localgit.ErrDiverged) ||
			errors.Is(err, localgit.ErrUpToDate) {
			return zero, badRequest(err.Error())
		}
		return zero, fmt.Errorf("pull ff-only: %w", err)
	}
	return PullFFResult{OK: true, Branch: branch, Path: abs}, nil
}
```

#### Commands.RequireMappedPath (method)

```go
func (c *Commands) RequireMappedPath(ctx context.Context, doc config.File, projectID, repoPath string) (string, error) {
	s := c.Service
	if s == nil || s.Local == nil {
		return "", fmt.Errorf("local git inspector missing")
	}
	projectID = strings.TrimSpace(projectID)
	repoPath = strings.TrimSpace(repoPath)
	if projectID == "" || repoPath == "" {
		return "", badRequest("project_id and path are required")
	}
	p, ok := FindProject(doc.Projects, projectID)
	if !ok {
		return "", badRequest("unknown project")
	}
	abs, err := localgit.ExpandPath(repoPath)
	if err != nil {
		return "", badRequest(fmt.Sprintf("path: %v", err))
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	var disc localgit.Discovery
	if len(doc.Local.Roots) > 0 {
		disc = s.Local.ScanRoots(ctx, doc.Local.Roots)
	}
	local := s.attachLocal(ctx, p, disc, projectLabelsByKey(doc.Projects), false, time.Duration(doc.EffectiveFetchSeconds())*time.Second)
	if local == nil || !local.Mapped {
		return "", badRequest("project has no mapped local checkout")
	}
	if !repoPathAllowed(local, abs) {
		return "", badRequest("path is not a mapped checkout for this project")
	}
	return abs, nil
}
```

#### SyncInvestigationRequest (type)

```go
type SyncInvestigationRequest struct {
	ProjectID string `json:"project_id"`
	Branch    string `json:"branch"`
	RepoPath  string `json:"repo_path"`
}
```

#### SyncInvestigation (type)

```go
type SyncInvestigation struct {
	ProjectID string                  `json:"project_id"`
	Result    localgit.SyncInspection `json:"result"`
}
```

#### Commands.InvestigateSync (method)

```go
func (c *Commands) InvestigateSync(ctx context.Context, doc config.File, req SyncInvestigationRequest) (SyncInvestigation, error) {
	var zero SyncInvestigation
	if c == nil || c.Service == nil || c.Local == nil {
		return zero, fmt.Errorf("local git inspector missing")
	}
	projectID := strings.TrimSpace(req.ProjectID)
	branch := strings.TrimSpace(req.Branch)
	repoPath := strings.TrimSpace(req.RepoPath)
	if projectID == "" || branch == "" || repoPath == "" {
		return zero, badRequest("project_id, branch, and repo_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return zero, badRequest(err.Error())
	}
	abs, err := c.RequireMappedPath(ctx, doc, projectID, repoPath)
	if err != nil {
		return zero, err
	}
	result, err := c.Local.InspectSync(ctx, abs, branch)
	if err != nil {
		return zero, fmt.Errorf("investigate sync: %w", err)
	}
	return SyncInvestigation{
		ProjectID: projectID,
		Result:    result,
	}, nil
}
```

### Private one-hop bodies

#### Service.annotateContentOnDefault (method)

```go
func (s *Service) annotateContentOnDefault(ctx context.Context, row *board.ProjectSummary) {
	if s == nil || s.Local == nil || row == nil || row.Local == nil || !row.Local.Mapped {
		return
	}
	if !row.RemoteNamesOK {
		return
	}

	remote := make(map[string]struct{}, len(row.RemoteNames))
	for _, name := range row.RemoteNames {
		name = strings.TrimSpace(name)
		if name != "" {
			remote[name] = struct{}{}
		}
	}

	defaultName := strings.TrimSpace(row.Local.DefaultBranch)
	for _, b := range row.Branches {
		if b.Default {
			if n := strings.TrimSpace(b.Name); n != "" {
				defaultName = n
			}
			break
		}
	}
	if defaultName == "" {
		defaultName = "main"
	}

	repoPath := strings.TrimSpace(row.Local.Path)
	if repoPath == "" {
		return
	}

	for i := range row.Local.Worktrees {
		wt := &row.Local.Worktrees[i]
		branch := strings.TrimSpace(wt.Branch)
		if branch == "" || wt.Detached || wt.Bare || wt.Dirty {
			continue
		}
		if branch == defaultName {
			continue
		}
		if _, onRemote := remote[branch]; onRemote {
			continue
		}
		checkPath := strings.TrimSpace(wt.Path)
		if checkPath == "" {
			checkPath = repoPath
		}
		ok, _, err := s.Local.ContentOnDefault(ctx, checkPath, branch, defaultName)
		if err != nil || !ok {
			continue
		}
		wt.ContentOnDefault = true
	}
}
```

#### Service.attachLocal (method)

```go
func (s *Service) attachLocal(ctx context.Context, p config.Project, disc localgit.Discovery, labelByKey map[string]string, fresh bool, fetchTTL time.Duration) *board.LocalStatus {
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
```

#### Service.summarize (method)

```go
func (s *Service) summarize(ctx context.Context, p config.Project, opts remotegit.SummaryOpts) board.ProjectSummary {
	switch p.Host {
	case config.HostGitHub:
		if s.GitHub == nil {
			return board.ProjectSummary{
				ID: p.ID, Label: p.Label, Host: string(p.Host), Path: p.Path, Org: forgeOrg(p.Path),
				OpenURL: p.OpenURL(), Error: "github client missing",
			}
		}
		row, _ := s.GitHub.ProjectSummary(ctx, p, opts)
		return row
	default:
		if s.GitLab == nil {
			return board.ProjectSummary{
				ID: p.ID, Label: p.Label, Host: string(p.Host), Path: p.Path, Org: forgeOrg(p.Path),
				OpenURL: p.OpenURL(), Error: "gitlab client missing",
			}
		}
		row, _ := s.GitLab.ProjectSummary(ctx, p, opts)
		return row
	}
}
```

#### badRequest (func)

```go
func badRequest(msg string) error {
	return BadRequestError{Msg: msg}
}
```

#### filterHiddenBranches (func)

```go
func filterHiddenBranches(branches []board.BranchRef, patterns []string) []board.BranchRef {
	if len(branches) == 0 || len(patterns) == 0 {
		return branches
	}
	compiled := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		compiled = append(compiled, p)
	}
	if len(compiled) == 0 {
		return branches
	}

	out := make([]board.BranchRef, 0, len(branches))
	for _, b := range branches {
		if branchHidden(b.Name, compiled) {
			continue
		}
		out = append(out, b)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
```

#### findSafeWorktree (func)

```go
func findSafeWorktree(local *board.LocalStatus, branch, absPath string) (board.LocalWorktree, bool) {
	if local == nil {
		return board.LocalWorktree{}, false
	}
	branch = strings.TrimSpace(branch)
	for _, wt := range local.Worktrees {
		if strings.TrimSpace(wt.Branch) != branch {
			continue
		}
		wtPath := filepath.Clean(wt.Path)
		if resolved, err := filepath.EvalSymlinks(wtPath); err == nil {
			wtPath = resolved
		}
		if wtPath != absPath {
			continue
		}
		if wt.PruneHint != board.PruneSafe {
			continue
		}
		wt.Path = wtPath
		return wt, true
	}
	return board.LocalWorktree{}, false
}
```

#### projectLabelsByKey (func)

```go
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
```

#### repoPathAllowed (func)

```go
func repoPathAllowed(local *board.LocalStatus, abs string) bool {
	if local == nil {
		return false
	}
	check := func(c string) bool {
		c = strings.TrimSpace(c)
		if c == "" {
			return false
		}
		path := filepath.Clean(c)
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
		return path == abs
	}
	if check(local.Path) {
		return true
	}
	for _, wt := range local.Worktrees {
		if check(wt.Path) {
			return true
		}
	}
	for _, app := range local.Appearances {
		if check(app.Path) {
			return true
		}
		for _, wt := range app.Worktrees {
			if check(wt.Path) {
				return true
			}
		}
	}
	return false
}
```


## ./internal/llm
- package: `llm`
- packageDoc: Package llm provides an OpenAI-compatible chat completions client.
- hasMain: false
- jsonTags: false
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: true
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- mechanicalRole: adapter
- mechanicalConfidence: 0.80
- mechanicalEvidence: imports_net_http, exports_client_surface, not_server
- exportedDecls: Client
- exportedFuncs: New
- exportedMethods: Client.Chat, Client.Enabled
- unexportedDecls: (none)
- unexportedFuncs: firstNonEmpty
- unexportedMethods: (none)
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/llm/client.go

### Exported bodies

#### Client (type)

```go
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}
```

#### New (func)

```go
func New(cfg config.LLM) *Client {
	return &Client{
		BaseURL: strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		APIKey:  strings.TrimSpace(cfg.APIKey),
		Model:   firstNonEmpty(cfg.Model, config.DefaultLLMModel),
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}
```

#### Client.Enabled (method)

```go
func (c *Client) Enabled() bool {
	return c != nil && strings.TrimSpace(c.BaseURL) != ""
}
```

#### Client.Chat (method)

```go
func (c *Client) Chat(ctx context.Context, system, user string) (content, model string, err error) {
	if !c.Enabled() {
		return "", "", fmt.Errorf("llm: base_url is not set")
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 120 * time.Second}
	}
	body := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.2,
	}
	rawBody, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("llm marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(rawBody))
	if err != nil {
		return "", "", fmt.Errorf("llm request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("llm chat: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", "", fmt.Errorf("llm read: %w", err)
	}
	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf("llm %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Model string `json:"model"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", "", fmt.Errorf("llm parse: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", "", fmt.Errorf("llm: empty choices")
	}
	return parsed.Choices[0].Message.Content, firstNonEmpty(parsed.Model, c.Model), nil
}
```

### Private one-hop bodies

#### firstNonEmpty (func)

```go
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
```


## ./internal/localgit
- package: `localgit`
- hasMain: false
- jsonTags: true
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- mechanicalRole: unknown
- mechanicalConfidence: 0.00
- exportedDecls: BranchSync, Checkout, Discovery, ErrDirtyTree, ErrDiverged, ErrInvalidBranch, ErrMissingBranch, ErrUpToDate, Inspector, OriginFetchCache, RemoteRef, RoleStandalone, RoleSubmodule, Status, SyncCommit, SyncInspection, Worktree
- exportedFuncs: DisplayID, ExpandPath, FillCheckoutMeta, InvalidateOriginSync, NewInspector, NewOriginFetchCache, ParseRemoteURL, PickPrimary, TrackKey, ValidateBranchName
- exportedMethods: Discovery.Appearances, Inspector.CommonGitDir, Inspector.ContentOnDefault, Inspector.EnrichOriginSync, Inspector.FetchOrigin, Inspector.FetchOriginCached, Inspector.InspectPath, Inspector.InspectSync, Inspector.OriginRemote, Inspector.PullFFOnly, Inspector.RemoveSafeCheckout, Inspector.ScanRoots, Inspector.Superproject, OriginFetchCache.Clear, OriginFetchCache.MarkSuccess, OriginFetchCache.NeedsFetch, OriginFetchCache.SetNow
- unexportedDecls: branchNameOK, exitCoder, maxScanDepth, originCompareParallel, syncCommitLimit
- unexportedFuncs: commandExitCode, homeShortPath, isWorkingTreeRoot, parseLeftRightCount, pathDepth, pickPrimary, porcelainFiles, remoteFromHostPath, resolvePath, summarize
- unexportedMethods: Inspector.defaultBranch, Inspector.ensureBranchFFFromOrigin, Inspector.ensureDefaultReadyForSafePrune, Inspector.exactTagAtHEAD, Inspector.git, Inspector.indexCheckout, Inspector.inspectWorktree, Inspector.isGitDir, Inspector.leftRight, Inspector.listRefShortNames, Inspector.listWorktrees, Inspector.resetUnusedBranchToOrigin, Inspector.resolveContentBaseline, Inspector.revParse, Inspector.syncCommits, Inspector.updateSubmodules, Inspector.worktreeOnBranch, OriginFetchCache.clock
- errorTypes: ErrDirtyTree, ErrDiverged, ErrInvalidBranch, ErrMissingBranch, ErrUpToDate
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/appearance.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/content.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/fetch.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/path.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/prune.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/pull.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/remote.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/scan.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/status.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/localgit/sync.go

### Exported bodies

#### Checkout (type)

```go
type Checkout struct {
	Path           string // absolute preferred worktree path
	CommonGitDir   string
	Superproject   string // absolute parent working tree when a submodule; empty if standalone
	Role           string // standalone | submodule
	RelPath        string // path relative to Superproject when submodule
	ParentBasename string // basename(Superproject) when submodule
}
```

#### Discovery.Appearances (method)

```go
func (d Discovery) Appearances(host, repoPath string) []Checkout {
	if d.ByKey == nil {
		return nil
	}
	out := d.ByKey[TrackKey(host, repoPath)]
	if len(out) == 0 {
		return nil
	}
	cp := make([]Checkout, len(out))
	copy(cp, out)
	return cp
}
```

#### PickPrimary (func)

```go
func PickPrimary(localPath string, checkouts []Checkout) (Checkout, bool) {
	localPath = strings.TrimSpace(localPath)
	if localPath != "" {
		abs, err := ExpandPath(localPath)
		if err == nil {
			for _, c := range checkouts {
				if filepath.Clean(c.Path) == filepath.Clean(abs) {
					return c, true
				}
			}
		}
		// Explicit override even when not in the scan list.
		return Checkout{Path: localPath, Role: RoleStandalone}, true
	}
	if len(checkouts) == 0 {
		return Checkout{}, false
	}
	sorted := make([]Checkout, len(checkouts))
	copy(sorted, checkouts)
	sort.SliceStable(sorted, func(i, j int) bool {
		si := sorted[i].Role == RoleStandalone
		sj := sorted[j].Role == RoleStandalone
		if si != sj {
			return si
		}
		return pathDepth(sorted[i].Path) < pathDepth(sorted[j].Path) ||
			(pathDepth(sorted[i].Path) == pathDepth(sorted[j].Path) && sorted[i].Path < sorted[j].Path)
	})
	return sorted[0], true
}
```

#### DisplayID (func)

```go
func DisplayID(c Checkout, parentLabel string) string {
	if c.Role == RoleSubmodule && c.Superproject != "" {
		label := strings.TrimSpace(parentLabel)
		if label == "" {
			label = c.ParentBasename
		}
		if label == "" {
			label = filepath.Base(c.Superproject)
		}
		rel := c.RelPath
		if rel == "" || rel == "." {
			rel = filepath.Base(c.Path)
		}
		return label + " → " + rel
	}
	return homeShortPath(c.Path)
}
```

#### FillCheckoutMeta (func)

```go
func FillCheckoutMeta(c *Checkout) {
	if c == nil {
		return
	}
	c.Path = filepath.Clean(c.Path)
	if strings.TrimSpace(c.Superproject) == "" {
		c.Role = RoleStandalone
		c.RelPath = ""
		c.ParentBasename = ""
		return
	}
	c.Superproject = filepath.Clean(c.Superproject)
	c.Role = RoleSubmodule
	c.ParentBasename = filepath.Base(c.Superproject)

	child := c.Path
	parent := c.Superproject
	if resolved, err := filepath.EvalSymlinks(child); err == nil {
		child = resolved
	}
	if resolved, err := filepath.EvalSymlinks(parent); err == nil {
		parent = resolved
		c.ParentBasename = filepath.Base(parent)
	}
	rel, err := filepath.Rel(parent, child)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		c.RelPath = filepath.Base(c.Path)
		return
	}
	c.RelPath = rel
}
```

#### Inspector.ContentOnDefault (method)

```go
func (in *Inspector) ContentOnDefault(ctx context.Context, repoPath, branch, defaultBranch string) (ok bool, reason string, err error) {
	if in == nil {
		return false, "", fmt.Errorf("inspector missing")
	}
	branch = strings.TrimSpace(branch)
	defaultBranch = strings.TrimSpace(defaultBranch)
	if branch == "" {
		return false, "", fmt.Errorf("branch is required")
	}
	if err := ValidateBranchName(branch); err != nil {
		return false, "", err
	}
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	if err := ValidateBranchName(defaultBranch); err != nil {
		return false, "", fmt.Errorf("default branch: %w", err)
	}
	if branch == defaultBranch {
		return false, "branch is default", nil
	}

	abs, err := ExpandPath(repoPath)
	if err != nil {
		return false, "", fmt.Errorf("repo path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	baseline, baselineReason, okBaseline := in.resolveContentBaseline(ctx, abs, defaultBranch)
	if !okBaseline {
		return false, baselineReason, nil
	}

	if _, err := in.git(ctx, abs, "merge-base", baseline, branch); err != nil {
		return false, "unrelated histories", nil
	}

	if _, err := in.git(ctx, abs, "merge-base", "--is-ancestor", branch, baseline); err == nil {
		return true, "ancestor of " + baseline, nil
	} else if code, hasCode := commandExitCode(err); hasCode && code == 1 {
		// Not an ancestor; fall through to tip-tree comparison.
	} else {
		return false, "ancestor check: " + err.Error(), nil
	}

	if _, err := in.git(ctx, abs, "diff", "--quiet", baseline, branch); err == nil {
		return true, "identical trees vs " + baseline, nil
	} else if code, hasCode := commandExitCode(err); hasCode && code == 1 {
		return false, "trees differ from " + baseline, nil
	}
	return false, "diff: " + err.Error(), nil
}
```

#### OriginFetchCache (type)

```go
type OriginFetchCache struct {
	mu  sync.Mutex
	now func() time.Time
	at  map[string]time.Time
}
```

#### NewOriginFetchCache (func)

```go
func NewOriginFetchCache() *OriginFetchCache {
	return &OriginFetchCache{
		now: time.Now,
		at:  map[string]time.Time{},
	}
}
```

#### OriginFetchCache.SetNow (method)

```go
func (c *OriginFetchCache) SetNow(now func() time.Time) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if now == nil {
		c.now = time.Now
		return
	}
	c.now = now
}
```

#### OriginFetchCache.Clear (method)

```go
func (c *OriginFetchCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = map[string]time.Time{}
}
```

#### OriginFetchCache.NeedsFetch (method)

```go
func (c *OriginFetchCache) NeedsFetch(commonDir string, ttl time.Duration, fresh bool) bool {
	commonDir = filepath.Clean(commonDir)
	if commonDir == "" {
		return true
	}
	if c == nil || ttl <= 0 || fresh {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	at, ok := c.at[commonDir]
	if !ok {
		return true
	}
	return c.clock().Sub(at) >= ttl
}
```

#### OriginFetchCache.MarkSuccess (method)

```go
func (c *OriginFetchCache) MarkSuccess(commonDir string) {
	if c == nil {
		return
	}
	commonDir = filepath.Clean(commonDir)
	if commonDir == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.at == nil {
		c.at = map[string]time.Time{}
	}
	c.at[commonDir] = c.clock()
}
```

#### Inspector.FetchOrigin (method)

```go
func (in *Inspector) FetchOrigin(ctx context.Context, repoPath string) error {
	if in == nil {
		return fmt.Errorf("inspector missing")
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("expand path: %w", err)
	}
	if _, err := in.git(ctx, abs, "fetch", "--prune", "origin"); err != nil {
		return fmt.Errorf("fetch --prune origin: %w", err)
	}
	return nil
}
```

#### Inspector.FetchOriginCached (method)

```go
func (in *Inspector) FetchOriginCached(ctx context.Context, repoPath string, ttl time.Duration, fresh bool, cache *OriginFetchCache) error {
	if in == nil {
		return fmt.Errorf("inspector missing")
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("expand path: %w", err)
	}
	common, err := in.CommonGitDir(ctx, abs)
	if err != nil || common == "" {
		common = abs
	}
	common = filepath.Clean(common)
	if !cache.NeedsFetch(common, ttl, fresh) {
		return nil
	}
	if err := in.FetchOrigin(ctx, abs); err != nil {
		return err
	}
	cache.MarkSuccess(common)
	return nil
}
```

#### InvalidateOriginSync (func)

```go
func InvalidateOriginSync(st *Status) {
	if st == nil {
		return
	}
	st.OriginSync = nil
	st.DefaultAhead = 0
	st.DefaultBehind = 0
}
```

#### ExpandPath (func)

```go
func ExpandPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("empty path")
	}
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		if p == "~" {
			p = home
		} else {
			p = filepath.Join(home, p[2:])
		}
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("abs %s: %w", p, err)
	}
	return abs, nil
}
```

#### TrackKey (func)

```go
func TrackKey(host, path string) string {
	return strings.ToLower(strings.TrimSpace(host)) + ":" + strings.ToLower(strings.Trim(path, "/"))
}
```

#### Inspector.RemoveSafeCheckout (method)

```go
func (in *Inspector) RemoveSafeCheckout(ctx context.Context, worktreePath, branch, defaultBranch string) error {
	if in == nil {
		return fmt.Errorf("inspector missing")
	}
	branch = strings.TrimSpace(branch)
	defaultBranch = strings.TrimSpace(defaultBranch)
	if branch == "" {
		return fmt.Errorf("branch is required")
	}
	if err := ValidateBranchName(branch); err != nil {
		return err
	}
	if defaultBranch == "" {
		defaultBranch = "main"
	}
	if err := ValidateBranchName(defaultBranch); err != nil {
		return fmt.Errorf("default branch: %w", err)
	}
	if branch == defaultBranch {
		return fmt.Errorf("refusing to remove default branch %q", branch)
	}

	abs, err := ExpandPath(worktreePath)
	if err != nil {
		return fmt.Errorf("worktree path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)

	trees, err := in.listWorktrees(ctx, abs)
	if err != nil {
		return fmt.Errorf("list worktrees: %w", err)
	}

	var target *Worktree
	var main *Worktree
	for i := range trees {
		wt := &trees[i]
		path := filepath.Clean(wt.Path)
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
		wt.Path = path
		if wt.Main {
			main = wt
		}
		if path == abs {
			target = wt
		}
	}
	if target == nil {
		return fmt.Errorf("worktree %s not found", abs)
	}
	if target.Bare {
		return fmt.Errorf("refusing to remove bare worktree")
	}
	if target.Detached {
		return fmt.Errorf("refusing to remove detached HEAD checkout")
	}
	if target.Locked {
		return fmt.Errorf("refusing to remove locked worktree")
	}
	if target.Branch != branch {
		if target.Branch == "" {
			return fmt.Errorf("worktree branch unknown; expected %q", branch)
		}
		return fmt.Errorf("worktree is on %q, not %q", target.Branch, branch)
	}

	detail, err := in.inspectWorktree(ctx, abs, target.Main)
	if err != nil {
		return fmt.Errorf("re-check worktree: %w", err)
	}
	if detail.Dirty {
		return fmt.Errorf("%w at %s; commit or stash before remove", ErrDirtyTree, abs)
	}

	if target.Main {
		cur, err := in.git(ctx, abs, "branch", "--show-current")
		if err != nil {
			return fmt.Errorf("current branch: %w", err)
		}
		if got := strings.TrimSpace(string(cur)); got != "" && got != branch {
			return fmt.Errorf("checkout is on %q, not %q", got, branch)
		}
		// Land on an up-to-date default instead of a stale tip after switch.
		if err := in.ensureDefaultReadyForSafePrune(ctx, abs, defaultBranch); err != nil {
			return fmt.Errorf("update default branch %s before remove: %w", defaultBranch, err)
		}
		if _, err := in.git(ctx, abs, "switch", defaultBranch); err != nil {
			return fmt.Errorf("switch to %s: %w", defaultBranch, err)
		}
		if _, err := in.git(ctx, abs, "branch", "-D", branch); err != nil {
			return fmt.Errorf("delete branch %s: %w", branch, err)
		}
		return nil
	}

	if main == nil {
		return fmt.Errorf("main worktree not found for %s", abs)
	}
	if _, err := in.git(ctx, main.Path, "worktree", "remove", abs); err != nil {
		return fmt.Errorf("worktree remove: %w", err)
	}
	if _, err := in.git(ctx, main.Path, "branch", "-D", branch); err != nil {
		return fmt.Errorf("worktree removed at %s; delete branch %s: %w", abs, branch, err)
	}
	return nil
}
```

#### ValidateBranchName (func)

```go
func ValidateBranchName(branch string) error {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return fmt.Errorf("%w: empty", ErrInvalidBranch)
	}
	if strings.EqualFold(branch, "HEAD") {
		return fmt.Errorf("%w: %q", ErrInvalidBranch, branch)
	}
	if strings.HasPrefix(branch, "-") || strings.HasPrefix(branch, ".") {
		return fmt.Errorf("%w: %q", ErrInvalidBranch, branch)
	}
	if strings.Contains(branch, "..") || strings.Contains(branch, "@{") {
		return fmt.Errorf("%w: %q", ErrInvalidBranch, branch)
	}
	if !branchNameOK.MatchString(branch) {
		return fmt.Errorf("%w: %q", ErrInvalidBranch, branch)
	}
	if len(branch) > 255 {
		return fmt.Errorf("%w: too long", ErrInvalidBranch)
	}
	return nil
}
```

#### Inspector.PullFFOnly (method)

```go
func (in *Inspector) PullFFOnly(ctx context.Context, repoPath, branch string) error {
	if in == nil {
		return fmt.Errorf("inspector missing")
	}
	branch = strings.TrimSpace(branch)
	if err := ValidateBranchName(branch); err != nil {
		return err
	}

	abs, err := ExpandPath(repoPath)
	if err != nil {
		return fmt.Errorf("repo path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)
	if !in.isGitDir(ctx, abs) {
		return fmt.Errorf("not a git repository: %s", abs)
	}

	localRef := "refs/heads/" + branch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", localRef); err != nil {
		return fmt.Errorf("%w: %q", ErrMissingBranch, branch)
	}

	checkout, checkedOut := in.worktreeOnBranch(ctx, abs, branch)
	if checkedOut {
		detail, err := in.inspectWorktree(ctx, checkout, false)
		if err != nil {
			return fmt.Errorf("inspect checkout: %w", err)
		}
		if detail.Dirty {
			// Submodule working trees often look dirty when they lag the parent tip.
			if syncErr := in.updateSubmodules(ctx, checkout); syncErr == nil {
				detail, err = in.inspectWorktree(ctx, checkout, false)
				if err != nil {
					return fmt.Errorf("inspect checkout: %w", err)
				}
			}
		}
		if detail.Dirty {
			return fmt.Errorf("%w at %s; commit or stash before pull", ErrDirtyTree, checkout)
		}
	}

	// Refresh origin/<branch> before comparing / merging.
	fetchDir := abs
	if checkedOut {
		fetchDir = checkout
	}
	if _, err := in.git(ctx, fetchDir, "fetch", "origin", branch); err != nil {
		return fmt.Errorf("fetch origin %s: %w", branch, err)
	}

	remoteRef := "refs/remotes/origin/" + branch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", remoteRef); err != nil {
		return fmt.Errorf("missing %s after fetch", remoteRef)
	}
	ahead, behind, ok := in.leftRight(ctx, abs, remoteRef, localRef)
	if !ok {
		return fmt.Errorf("could not compare %s to origin/%s", branch, branch)
	}
	if behind == 0 {
		return fmt.Errorf("%w: %q", ErrUpToDate, branch)
	}
	if ahead > 0 {
		return fmt.Errorf("%w: %q (%d ahead, %d behind)", ErrDiverged, branch, ahead, behind)
	}

	if checkedOut {
		if _, err := in.git(ctx, checkout, "merge", "--ff-only", "origin/"+branch); err != nil {
			return fmt.Errorf("ff-only merge origin/%s: %w", branch, err)
		}
		if err := in.updateSubmodules(ctx, checkout); err != nil {
			return fmt.Errorf("after fast-forward: %w", err)
		}
		return nil
	}

	spec := branch + ":" + branch
	if _, err := in.git(ctx, abs, "fetch", "origin", spec); err != nil {
		return fmt.Errorf("fetch origin %s: %w", spec, err)
	}
	return nil
}
```

#### RemoteRef (type)

```go
type RemoteRef struct {
	Host string // github | gitlab
	Path string // owner/repo (no .git)
}
```

#### ParseRemoteURL (func)

```go
func ParseRemoteURL(raw string) (RemoteRef, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return RemoteRef{}, false
	}
	raw = strings.TrimSuffix(raw, ".git")

	// git@host:owner/repo or ssh://git@host/owner/repo
	if strings.HasPrefix(raw, "git@") {
		rest := strings.TrimPrefix(raw, "git@")
		host, path, ok := strings.Cut(rest, ":")
		if !ok {
			return RemoteRef{}, false
		}
		return remoteFromHostPath(host, path)
	}
	if strings.HasPrefix(raw, "ssh://") {
		u, err := url.Parse(raw)
		if err != nil {
			return RemoteRef{}, false
		}
		return remoteFromHostPath(u.Host, strings.TrimPrefix(u.Path, "/"))
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err != nil {
			return RemoteRef{}, false
		}
		return remoteFromHostPath(u.Host, strings.TrimPrefix(u.Path, "/"))
	}
	return RemoteRef{}, false
}
```

#### Discovery (type)

```go
type Discovery struct {
	// ByKey maps "github:owner/repo" → checkouts (deduped by common git dir).
	ByKey map[string][]Checkout
}
```

#### Inspector.ScanRoots (method)

```go
func (in *Inspector) ScanRoots(ctx context.Context, roots []string) Discovery {
	d := Discovery{ByKey: map[string][]Checkout{}}
	seenCommon := map[string]struct{}{} // common git dir already indexed
	for _, root := range roots {
		abs, err := ExpandPath(root)
		if err != nil {
			continue
		}
		st, err := os.Stat(abs)
		if err != nil || !st.IsDir() {
			continue
		}
		walkErr := filepath.WalkDir(abs, func(path string, de os.DirEntry, entryErr error) error {
			if entryErr != nil {
				// Skip individual unreadable entries but propagate ctx cancel.
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if !de.IsDir() {
				return nil
			}
			name := de.Name()
			if name == "node_modules" || name == "vendor" || name == ".git" {
				if name == ".git" {
					return filepath.SkipDir
				}
				return filepath.SkipDir
			}
			rel, err := filepath.Rel(abs, path)
			if err != nil {
				return nil
			}
			if rel != "." {
				depth := strings.Count(rel, string(os.PathSeparator)) + 1
				if depth > maxScanDepth {
					return filepath.SkipDir
				}
			}
			gitMeta := filepath.Join(path, ".git")
			if _, err := os.Stat(gitMeta); err != nil {
				return nil
			}
			in.indexCheckout(ctx, path, d, seenCommon)
			// Keep walking so nested clones / submodules are indexed too.
			return nil
		})
		if walkErr != nil && ctx.Err() != nil {
			// Context canceled; stop processing further roots.
			return d
		}
	}
	return d
}
```

#### Inspector.Superproject (method)

```go
func (in *Inspector) Superproject(ctx context.Context, dir string) string {
	out, err := in.git(ctx, dir, "rev-parse", "--show-superproject-working-tree")
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(out))
	if s == "" || s == "." {
		return ""
	}
	return filepath.Clean(s)
}
```

#### Worktree (type)

```go
type Worktree struct {
	Path     string `json:"path"`
	Branch   string `json:"branch,omitempty"`
	Tag      string `json:"tag,omitempty"` // Exact tag when Detached and HEAD is tagged.
	Detached bool   `json:"detached,omitempty"`
	Bare     bool   `json:"bare,omitempty"`
	Locked   bool   `json:"locked,omitempty"`
	Prunable bool   `json:"prunable,omitempty"`
	Main     bool   `json:"main,omitempty"`
	Dirty    bool   `json:"dirty,omitempty"`
	Ahead    int    `json:"ahead,omitempty"`
	Behind   int    `json:"behind,omitempty"`
	Upstream string `json:"upstream,omitempty"`
}
```

#### Status (type)

```go
type Status struct {
	Mapped        bool         `json:"mapped"`
	Path          string       `json:"path,omitempty"`
	Error         string       `json:"error,omitempty"`
	Branch        string       `json:"branch,omitempty"`
	Tag           string       `json:"tag,omitempty"` // Exact tag when Detached and HEAD is tagged.
	Detached      bool         `json:"detached,omitempty"`
	Dirty         bool         `json:"dirty,omitempty"`
	Ahead         int          `json:"ahead,omitempty"`
	Behind        int          `json:"behind,omitempty"`
	Upstream      string       `json:"upstream,omitempty"`
	DefaultBranch string       `json:"default_branch,omitempty"`
	DefaultBehind int          `json:"default_behind,omitempty"`
	DefaultAhead  int          `json:"default_ahead,omitempty"`
	OriginSync    []BranchSync `json:"origin_sync,omitempty"`
	Worktrees     []Worktree   `json:"worktrees,omitempty"`
}
```

#### BranchSync (type)

```go
type BranchSync struct {
	Name   string `json:"name"`
	Ahead  int    `json:"ahead,omitempty"`  // local has commits origin lacks
	Behind int    `json:"behind,omitempty"` // origin has commits local lacks
}
```

#### Inspector (type)

```go
type Inspector struct {
	Run cliexec.Exec
}
```

#### NewInspector (func)

```go
func NewInspector(run cliexec.Exec) *Inspector {
	if run == nil {
		panic("localgit.NewInspector: Exec is required")
	}
	return &Inspector{Run: run}
}
```

#### Inspector.InspectPath (method)

```go
func (in *Inspector) InspectPath(ctx context.Context, path string) Status {
	abs, err := ExpandPath(path)
	if err != nil {
		return Status{Mapped: true, Error: err.Error()}
	}
	if _, err := os.Stat(abs); err != nil {
		return Status{Mapped: true, Path: abs, Error: fmt.Sprintf("path missing: %v", err)}
	}
	if !in.isGitDir(ctx, abs) {
		return Status{Mapped: true, Path: abs, Error: "not a git repository"}
	}

	trees, err := in.listWorktrees(ctx, abs)
	if err != nil {
		// Fall back to single-tree inspect.
		wt, wtErr := in.inspectWorktree(ctx, abs, true)
		if wtErr != nil {
			return Status{Mapped: true, Path: abs, Error: err.Error()}
		}
		return summarize(abs, []Worktree{wt})
	}
	for i := range trees {
		if trees[i].Bare {
			continue
		}
		detail, detailErr := in.inspectWorktree(ctx, trees[i].Path, trees[i].Main)
		if detailErr != nil {
			// Fail closed for prune: never leave Dirty as a false zero-value.
			trees[i].Dirty = true
			if trees[i].Branch == "" {
				trees[i].Branch = "unknown"
			}
			continue
		}
		detail.Main = trees[i].Main
		detail.Bare = trees[i].Bare
		detail.Locked = trees[i].Locked
		detail.Prunable = trees[i].Prunable
		if trees[i].Branch != "" && detail.Branch == "" {
			detail.Branch = trees[i].Branch
		}
		trees[i] = detail
	}
	return summarize(abs, trees)
}
```

#### Inspector.EnrichOriginSync (method)

```go
func (in *Inspector) EnrichOriginSync(ctx context.Context, st *Status) {
	if in == nil || st == nil || !st.Mapped || st.Path == "" || st.Error != "" {
		return
	}
	if def, ok := in.defaultBranch(ctx, st.Path); ok {
		st.DefaultBranch = def
	}
	localNames, err := in.listRefShortNames(ctx, st.Path, "refs/heads/")
	if err != nil {
		if st.Error == "" {
			st.Error = "list local branches: " + err.Error()
		}
		return
	}
	remoteNames, err := in.listRefShortNames(ctx, st.Path, "refs/remotes/origin/")
	if err != nil {
		if st.Error == "" {
			st.Error = "list origin branches: " + err.Error()
		}
		return
	}
	remoteSet := make(map[string]struct{}, len(remoteNames))
	for _, n := range remoteNames {
		n = strings.TrimPrefix(n, "origin/")
		if n == "" || n == "HEAD" {
			continue
		}
		remoteSet[n] = struct{}{}
	}
	var names []string
	for _, name := range localNames {
		if _, ok := remoteSet[name]; ok {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		st.OriginSync = nil
		return
	}

	syncs := make([]BranchSync, len(names))
	var wg sync.WaitGroup
	var failed atomic.Bool
	sem := make(chan struct{}, originCompareParallel)
	for i, name := range names {
		if ctx.Err() != nil {
			failed.Store(true)
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, name string) {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				failed.Store(true)
				return
			}
			ahead, behind, ok := in.leftRight(ctx, st.Path, "refs/remotes/origin/"+name, "refs/heads/"+name)
			if !ok {
				failed.Store(true)
				return
			}
			syncs[i] = BranchSync{Name: name, Ahead: ahead, Behind: behind}
		}(i, name)
	}
	wg.Wait()

	if failed.Load() {
		if st.Error == "" {
			st.Error = "could not compare local branches to origin"
		}
		st.OriginSync = nil
		return
	}
	out := make([]BranchSync, 0, len(syncs))
	for _, s := range syncs {
		if s.Name == "" {
			continue
		}
		out = append(out, s)
		if s.Name == st.DefaultBranch {
			st.DefaultAhead = s.Ahead
			st.DefaultBehind = s.Behind
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	st.OriginSync = out
}
```

#### Inspector.OriginRemote (method)

```go
func (in *Inspector) OriginRemote(ctx context.Context, dir string) (string, error) {
	out, err := in.git(ctx, dir, "remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
```

#### Inspector.CommonGitDir (method)

```go
func (in *Inspector) CommonGitDir(ctx context.Context, dir string) (string, error) {
	out, err := in.git(ctx, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return filepath.Clean(strings.TrimSpace(string(out))), nil
}
```

#### SyncCommit (type)

```go
type SyncCommit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
}
```

#### SyncInspection (type)

```go
type SyncInspection struct {
	Path           string       `json:"path"`
	Branch         string       `json:"branch"`
	CurrentBranch  string       `json:"current_branch,omitempty"`
	Upstream       string       `json:"upstream,omitempty"`
	LocalSHA       string       `json:"local_sha,omitempty"`
	RemoteSHA      string       `json:"remote_sha,omitempty"`
	MergeBase      string       `json:"merge_base,omitempty"`
	Relation       string       `json:"relation"`
	Dirty          bool         `json:"dirty"`
	DirtyFiles     []string     `json:"dirty_files,omitempty"`
	AheadCount     int          `json:"ahead_count"`
	BehindCount    int          `json:"behind_count"`
	Ahead          []SyncCommit `json:"ahead,omitempty"`
	Behind         []SyncCommit `json:"behind,omitempty"`
	CanFastForward bool         `json:"can_fast_forward"`
}
```

#### Inspector.InspectSync (method)

```go
func (in *Inspector) InspectSync(ctx context.Context, repoPath, branch string) (SyncInspection, error) {
	var out SyncInspection
	if in == nil {
		return out, fmt.Errorf("inspector missing")
	}
	branch = strings.TrimSpace(branch)
	if err := ValidateBranchName(branch); err != nil {
		return out, err
	}
	abs, err := ExpandPath(repoPath)
	if err != nil {
		return out, fmt.Errorf("expand path: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(abs); resolveErr == nil {
		abs = resolved
	}
	abs = filepath.Clean(abs)
	if !in.isGitDir(ctx, abs) {
		return out, fmt.Errorf("not a git repository: %s", abs)
	}

	out.Path = abs
	out.Branch = branch
	out.Upstream = "origin/" + branch
	if _, err := in.git(ctx, abs, "fetch", "--prune", "origin"); err != nil {
		return out, fmt.Errorf("fetch --prune origin: %w", err)
	}

	localRef := "refs/heads/" + branch
	remoteRef := "refs/remotes/origin/" + branch
	out.LocalSHA, err = in.revParse(ctx, abs, localRef)
	if err != nil {
		return out, fmt.Errorf("local branch %q: %w", branch, err)
	}
	out.RemoteSHA, err = in.revParse(ctx, abs, remoteRef)
	if err != nil {
		return out, fmt.Errorf("origin branch %q: %w", branch, err)
	}

	current, currentErr := in.git(ctx, abs, "rev-parse", "--abbrev-ref", "HEAD")
	if currentErr != nil {
		return out, fmt.Errorf("current branch: %w", currentErr)
	}
	out.CurrentBranch = strings.TrimSpace(string(current))
	status, statusErr := in.git(ctx, abs, "status", "--porcelain")
	if statusErr != nil {
		return out, fmt.Errorf("status: %w", statusErr)
	}
	if out.CurrentBranch == branch {
		out.DirtyFiles = porcelainFiles(string(status))
		out.Dirty = len(out.DirtyFiles) > 0
	}

	counts, countsErr := in.git(ctx, abs, "rev-list", "--left-right", "--count", remoteRef+"..."+localRef)
	if countsErr != nil {
		return out, fmt.Errorf("compare branch %q with origin: %w", branch, countsErr)
	}
	out.BehindCount, out.AheadCount, err = parseLeftRightCount(string(counts))
	if err != nil {
		return out, fmt.Errorf("compare branch %q with origin: %w", branch, err)
	}

	if base, baseErr := in.git(ctx, abs, "merge-base", localRef, remoteRef); baseErr == nil {
		out.MergeBase = strings.TrimSpace(string(base))
	} else {
		out.Relation = "unrelated"
		return out, nil
	}

	out.Behind, err = in.syncCommits(ctx, abs, localRef+".."+remoteRef)
	if err != nil {
		return out, fmt.Errorf("list commits behind %q: %w", branch, err)
	}
	out.Ahead, err = in.syncCommits(ctx, abs, remoteRef+".."+localRef)
	if err != nil {
		return out, fmt.Errorf("list commits ahead of %q: %w", branch, err)
	}

	switch {
	case out.AheadCount == 0 && out.BehindCount == 0:
		out.Relation = "up_to_date"
	case out.AheadCount == 0:
		out.Relation = "behind_only"
		out.CanFastForward = !out.Dirty
	case out.BehindCount == 0:
		out.Relation = "ahead_only"
	default:
		out.Relation = "diverged"
	}
	return out, nil
}
```

### Private one-hop bodies

#### Inspector.defaultBranch (method)

```go
func (in *Inspector) defaultBranch(ctx context.Context, dir string) (string, bool) {
	out, err := in.git(ctx, dir, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD")
	if err == nil {
		ref := strings.TrimSpace(string(out))
		ref = strings.TrimPrefix(ref, "refs/remotes/origin/")
		if ref != "" {
			return ref, true
		}
	}
	for _, name := range []string{"main", "master"} {
		if _, err := in.git(ctx, dir, "rev-parse", "--verify", "refs/remotes/origin/"+name); err == nil {
			return name, true
		}
	}
	return "", false
}
```

#### Inspector.ensureDefaultReadyForSafePrune (method)

```go
func (in *Inspector) ensureDefaultReadyForSafePrune(ctx context.Context, repoPath, branch string) error {
	err := in.ensureBranchFFFromOrigin(ctx, repoPath, branch)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrDiverged) {
		return err
	}
	if _, checkedOut := in.worktreeOnBranch(ctx, repoPath, branch); checkedOut {
		return err
	}
	return in.resetUnusedBranchToOrigin(ctx, repoPath, branch)
}
```

#### Inspector.git (method)

```go
func (in *Inspector) git(ctx context.Context, dir string, args ...string) ([]byte, error) {
	argv := append([]string{"-C", dir}, args...)
	return in.Run.Run(ctx, "git", argv...)
}
```

#### Inspector.indexCheckout (method)

```go
func (in *Inspector) indexCheckout(ctx context.Context, path string, d Discovery, seenCommon map[string]struct{}) {
	if !in.isGitDir(ctx, path) {
		return
	}
	url, err := in.OriginRemote(ctx, path)
	if err != nil {
		return
	}
	ref, ok := ParseRemoteURL(url)
	if !ok {
		return
	}
	key := TrackKey(ref.Host, ref.Path)
	common, err := in.CommonGitDir(ctx, path)
	if err != nil {
		common = path
	}
	common = filepath.Clean(common)
	if _, exists := seenCommon[common]; exists {
		return
	}

	// Prefer the main worktree when multiple checkouts share a common dir,
	// but never replace a working tree with the bare/common git directory.
	trees, listErr := in.listWorktrees(ctx, path)
	preferred := path
	if listErr == nil {
		for _, wt := range trees {
			if !wt.Main || wt.Bare || wt.Path == "" {
				continue
			}
			cand := filepath.Clean(wt.Path)
			if cand == common || !isWorkingTreeRoot(cand) {
				continue
			}
			preferred = cand
			break
		}
	}

	c := Checkout{
		Path:         preferred,
		CommonGitDir: common,
		Superproject: in.Superproject(ctx, preferred),
	}
	FillCheckoutMeta(&c)
	seenCommon[common] = struct{}{}
	d.ByKey[key] = append(d.ByKey[key], c)
}
```

#### Inspector.inspectWorktree (method)

```go
func (in *Inspector) inspectWorktree(ctx context.Context, dir string, isMain bool) (Worktree, error) {
	wt := Worktree{Path: dir, Main: isMain}

	branchOut, err := in.git(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return wt, err
	}
	branch := strings.TrimSpace(string(branchOut))
	if branch == "HEAD" {
		wt.Detached = true
		short, shortErr := in.git(ctx, dir, "rev-parse", "--short", "HEAD")
		if shortErr == nil {
			wt.Branch = strings.TrimSpace(string(short))
		}
		wt.Tag = in.exactTagAtHEAD(ctx, dir)
	} else {
		wt.Branch = branch
	}

	dirtyOut, err := in.git(ctx, dir, "status", "--porcelain")
	if err != nil {
		return wt, err
	}
	wt.Dirty = len(bytes.TrimSpace(dirtyOut)) > 0

	upOut, upErr := in.git(ctx, dir, "rev-parse", "--abbrev-ref", "@{upstream}")
	if upErr == nil {
		wt.Upstream = strings.TrimSpace(string(upOut))
		ahead, behind, ok := in.leftRight(ctx, dir, "@{upstream}", "HEAD")
		if ok {
			wt.Ahead = ahead
			wt.Behind = behind
		}
	}
	return wt, nil
}
```

#### Inspector.isGitDir (method)

```go
func (in *Inspector) isGitDir(ctx context.Context, dir string) bool {
	_, err := in.git(ctx, dir, "rev-parse", "--git-dir")
	return err == nil
}
```

#### Inspector.leftRight (method)

```go
func (in *Inspector) leftRight(ctx context.Context, dir, left, right string) (ahead, behind int, ok bool) {
	// git rev-list --left-right --count A...B → "<behind>\t<ahead>" relative to B vs A
	// With A=upstream and B=HEAD: left=commits reachable from A not B (behind), right=ahead.
	out, err := in.git(ctx, dir, "rev-list", "--left-right", "--count", left+"..."+right)
	if err != nil {
		return 0, 0, false
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0, false
	}
	behind, err1 := strconv.Atoi(parts[0])
	ahead, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return ahead, behind, true
}
```

#### Inspector.listRefShortNames (method)

```go
func (in *Inspector) listRefShortNames(ctx context.Context, dir, prefix string) ([]string, error) {
	out, err := in.git(ctx, dir, "for-each-ref", "--format=%(refname:short)", prefix)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	var names []string
	for sc.Scan() {
		n := strings.TrimSpace(sc.Text())
		if n == "" {
			continue
		}
		names = append(names, n)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return names, nil
}
```

#### Inspector.listWorktrees (method)

```go
func (in *Inspector) listWorktrees(ctx context.Context, dir string) ([]Worktree, error) {
	out, err := in.git(ctx, dir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var trees []Worktree
	var cur *Worktree
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			cur = nil
			continue
		}
		key, val, _ := strings.Cut(line, " ")
		switch key {
		case "worktree":
			wt := Worktree{Path: val, Main: len(trees) == 0}
			trees = append(trees, wt)
			cur = &trees[len(trees)-1]
		case "bare":
			if cur != nil {
				cur.Bare = true
			}
		case "detached":
			if cur != nil {
				cur.Detached = true
			}
		case "branch":
			if cur != nil {
				cur.Branch = strings.TrimPrefix(val, "refs/heads/")
			}
		case "locked":
			if cur != nil {
				cur.Locked = true
			}
		case "prunable":
			if cur != nil {
				cur.Prunable = true
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(trees) == 0 {
		return nil, fmt.Errorf("no worktrees listed")
	}
	return trees, nil
}
```

#### Inspector.resolveContentBaseline (method)

```go
func (in *Inspector) resolveContentBaseline(ctx context.Context, abs, defaultBranch string) (ref, reason string, ok bool) {
	originRef := "refs/remotes/origin/" + defaultBranch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", originRef); err == nil {
		return originRef, "", true
	}
	localRef := "refs/heads/" + defaultBranch
	if _, err := in.git(ctx, abs, "rev-parse", "--verify", localRef); err == nil {
		return localRef, "", true
	}
	return "", "missing baseline " + defaultBranch, false
}
```

#### Inspector.revParse (method)

```go
func (in *Inspector) revParse(ctx context.Context, dir, ref string) (string, error) {
	out, err := in.git(ctx, dir, "rev-parse", "--verify", ref)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(out))
	if value == "" {
		return "", fmt.Errorf("empty ref")
	}
	return value, nil
}
```

#### Inspector.syncCommits (method)

```go
func (in *Inspector) syncCommits(ctx context.Context, dir, rangeSpec string) ([]SyncCommit, error) {
	out, err := in.git(ctx, dir, "log", "--max-count", fmt.Sprint(syncCommitLimit), "--format=%h%x09%s", rangeSpec)
	if err != nil {
		return nil, err
	}
	var commits []SyncCommit
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		sha, subject, ok := strings.Cut(line, "\t")
		if !ok {
			commits = append(commits, SyncCommit{SHA: line})
			continue
		}
		commits = append(commits, SyncCommit{
			SHA:     strings.TrimSpace(sha),
			Subject: strings.TrimSpace(subject),
		})
	}
	return commits, nil
}
```

#### Inspector.updateSubmodules (method)

```go
func (in *Inspector) updateSubmodules(ctx context.Context, dir string) error {
	if _, err := in.git(ctx, dir, "submodule", "update", "--init", "--recursive"); err != nil {
		return fmt.Errorf("submodule update --init --recursive: %w", err)
	}
	return nil
}
```

#### Inspector.worktreeOnBranch (method)

```go
func (in *Inspector) worktreeOnBranch(ctx context.Context, dir, branch string) (path string, ok bool) {
	trees, err := in.listWorktrees(ctx, dir)
	if err != nil {
		return "", false
	}
	for _, wt := range trees {
		if wt.Bare || wt.Detached {
			continue
		}
		if wt.Branch == branch {
			path := filepath.Clean(wt.Path)
			if resolved, err := filepath.EvalSymlinks(path); err == nil {
				path = resolved
			}
			return path, true
		}
	}
	return "", false
}
```

#### OriginFetchCache.clock (method)

```go
func (c *OriginFetchCache) clock() time.Time {
	if c == nil || c.now == nil {
		return time.Now()
	}
	return c.now()
}
```

#### commandExitCode (func)

```go
func commandExitCode(err error) (int, bool) {
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode(), true
	}
	return 0, false
}
```

#### homeShortPath (func)

```go
func homeShortPath(abs string) string {
	abs = filepath.Clean(strings.TrimSpace(abs))
	if abs == "" {
		return abs
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return abs
	}
	home = filepath.Clean(home)
	if abs == home {
		return "~"
	}
	prefix := home + string(filepath.Separator)
	if strings.HasPrefix(abs, prefix) {
		return "~/" + abs[len(prefix):]
	}
	return abs
}
```

#### parseLeftRightCount (func)

```go
func parseLeftRightCount(value string) (behind, ahead int, err error) {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid count %q", strings.TrimSpace(value))
	}
	if _, err := fmt.Sscanf(parts[0], "%d", &behind); err != nil {
		return 0, 0, fmt.Errorf("invalid behind count %q: %w", parts[0], err)
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &ahead); err != nil {
		return 0, 0, fmt.Errorf("invalid ahead count %q: %w", parts[1], err)
	}
	return behind, ahead, nil
}
```

#### pathDepth (func)

```go
func pathDepth(p string) int {
	p = filepath.Clean(p)
	if p == "" || p == string(filepath.Separator) {
		return 0
	}
	return strings.Count(p, string(filepath.Separator))
}
```

#### porcelainFiles (func)

```go
func porcelainFiles(status string) []string {
	var files []string
	for _, line := range strings.Split(status, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}
```

#### remoteFromHostPath (func)

```go
func remoteFromHostPath(host, path string) (RemoteRef, bool) {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, ok := strings.Cut(host, ":"); ok { // host:port
		host = h
	}
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return RemoteRef{}, false
	}
	// Keep full path for nested GitLab groups (group/sub/repo).
	path = strings.Join(parts, "/")

	var forge string
	switch {
	case host == "github.com" || strings.HasSuffix(host, ".github.com"):
		forge = "github"
	case host == "gitlab.com" || strings.Contains(host, "gitlab"):
		forge = "gitlab"
	default:
		return RemoteRef{}, false
	}
	return RemoteRef{Host: forge, Path: path}, true
}
```

#### summarize (func)

```go
func summarize(path string, trees []Worktree) Status {
	st := Status{Mapped: true, Path: path, Worktrees: trees}
	primary := pickPrimary(trees)
	if primary == nil {
		return st
	}
	st.Path = primary.Path
	st.Branch = primary.Branch
	st.Tag = primary.Tag
	st.Detached = primary.Detached
	st.Dirty = primary.Dirty
	st.Ahead = primary.Ahead
	st.Behind = primary.Behind
	st.Upstream = primary.Upstream
	return st
}
```


## ./internal/observability
- package: `observability`
- packageDoc: Package observability bootstraps OTEL and writes error-only inference dumps (same pattern as content-pipelines).
- hasMain: false
- jsonTags: true
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: false
- importsGrpc: false
- importsOtel: true
- importsPrometheus: false
- mechanicalRole: observability
- mechanicalConfidence: 0.90
- mechanicalEvidence: imports_otel
- exportedDecls: InitConfig
- exportedFuncs: Init, Shutdown
- exportedMethods: failureDumpProcessor.ForceFlush, failureDumpProcessor.OnEnd, failureDumpProcessor.OnStart, failureDumpProcessor.Shutdown
- unexportedDecls: _, dumpedEvent, dumpedSpan, dumpedTrace, embeddedHTTPURL, failureDumpProcessor, traceDumpBuffer
- unexportedFuncs: attributesToMap, logf, newFailureDumpProcessor, pruneFailureDumps, redactURLsInText, sanitizeAttrValue, snapshotEvents, snapshotSpan, statusCodeString
- unexportedMethods: failureDumpProcessor.writeLocked
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/observability/failure_dump.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/observability/otel.go

### Exported bodies

#### failureDumpProcessor.OnStart (method)

```go
func (p *failureDumpProcessor) OnStart(context.Context, sdktrace.ReadWriteSpan) {}
```

#### failureDumpProcessor.OnEnd (method)

```go
func (p *failureDumpProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	if p == nil || s == nil {
		return
	}
	snap := snapshotSpan(s)
	traceID := snap.TraceID
	if traceID == "" {
		return
	}
	isRoot := snap.ParentSpanID == ""
	isError := s.Status().Code == codes.Error

	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.traces) > 256 {
		// Bound memory when root spans never end.
		for tid := range p.traces {
			delete(p.traces, tid)
			if len(p.traces) <= 128 {
				break
			}
		}
	}
	buf := p.traces[traceID]
	if buf == nil {
		buf = &traceDumpBuffer{}
		p.traces[traceID] = buf
	}
	buf.spans = append(buf.spans, snap)
	if isError {
		buf.hasError = true
	}
	if !isRoot {
		return
	}
	delete(p.traces, traceID)
	if !isError {
		return
	}
	p.writeLocked(dumpedTrace{
		TraceID:           traceID,
		DumpedAt:          time.Now().UTC(),
		Reason:            "root_span_error",
		RootName:          snap.Name,
		RootStatusMessage: snap.StatusMessage,
		Spans:             buf.spans,
	})
}
```

#### failureDumpProcessor.Shutdown (method)

```go
func (p *failureDumpProcessor) Shutdown(context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for tid, buf := range p.traces {
		if buf != nil && buf.hasError {
			p.writeLocked(dumpedTrace{
				TraceID:  tid,
				DumpedAt: time.Now().UTC(),
				Reason:   "shutdown_with_error_spans",
				Spans:    buf.spans,
			})
		}
		delete(p.traces, tid)
	}
	return nil
}
```

#### failureDumpProcessor.ForceFlush (method)

```go
func (p *failureDumpProcessor) ForceFlush(context.Context) error { return nil }
```

#### InitConfig (type)

```go
type InitConfig struct {
	ServiceName            string
	OTLPEndpoint           string
	FailureDumpDir         string
	FailureDumpMaxAgeHours int
	FailureDumpMaxFiles    int
}
```

#### Init (func)

```go
func Init(cfg InitConfig) (*sdktrace.TracerProvider, error) {
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "gitboard"
	}
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(1.0)),
	}
	if cfg.FailureDumpDir != "" {
		opts = append(opts, sdktrace.WithSpanProcessor(newFailureDumpProcessor(
			cfg.FailureDumpDir,
			cfg.FailureDumpMaxAgeHours,
			cfg.FailureDumpMaxFiles,
		)))
	}
	if cfg.OTLPEndpoint != "" {
		exporter, exportErr := otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if exportErr != nil {
			return nil, fmt.Errorf("otlp exporter: %w", exportErr)
		}
		opts = append(opts, sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(2*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		))
	}
	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp, nil
}
```

#### Shutdown (func)

```go
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}
```

### Private one-hop bodies

#### dumpedTrace (type)

```go
type dumpedTrace struct {
	TraceID           string       `json:"trace_id"`
	DumpedAt          time.Time    `json:"dumped_at"`
	Reason            string       `json:"reason"`
	RootName          string       `json:"root_name,omitempty"`
	RootStatusMessage string       `json:"root_status_message,omitempty"`
	SpanCount         int          `json:"span_count"`
	Spans             []dumpedSpan `json:"spans"`
}
```

#### failureDumpProcessor (type)

```go
type failureDumpProcessor struct {
	dir         string
	maxAgeHours int
	maxFiles    int
	mu          sync.Mutex
	traces      map[string]*traceDumpBuffer
}
```

#### failureDumpProcessor.writeLocked (method)

```go
func (p *failureDumpProcessor) writeLocked(doc dumpedTrace) {
	sort.Slice(doc.Spans, func(i, j int) bool {
		return doc.Spans[i].StartTime.Before(doc.Spans[j].StartTime)
	})
	doc.SpanCount = len(doc.Spans)
	if err := os.MkdirAll(p.dir, 0o750); err != nil {
		logf("failure dump: mkdir dir=%s: %v", p.dir, err)
		return
	}
	path := filepath.Join(p.dir, doc.TraceID+".json")
	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		logf("failure dump: marshal trace_id=%s: %v", doc.TraceID, err)
		return
	}
	if err := os.WriteFile(path, body, 0o640); err != nil {
		logf("failure dump: write path=%s trace_id=%s: %v", path, doc.TraceID, err)
		return
	}
	logf("Inference failure dump written path=%s trace_id=%s spans=%d", path, doc.TraceID, doc.SpanCount)
	if err := pruneFailureDumps(p.dir, p.maxAgeHours, p.maxFiles); err != nil {
		logf("prune dumps: %v", err)
	}
}
```

#### newFailureDumpProcessor (func)

```go
func newFailureDumpProcessor(dir string, maxAgeHours, maxFiles int) *failureDumpProcessor {
	return &failureDumpProcessor{
		dir:         dir,
		maxAgeHours: maxAgeHours,
		maxFiles:    maxFiles,
		traces:      make(map[string]*traceDumpBuffer),
	}
}
```

#### snapshotSpan (func)

```go
func snapshotSpan(s sdktrace.ReadOnlySpan) dumpedSpan {
	sc := s.SpanContext()
	parentID := ""
	if s.Parent().IsValid() {
		parentID = s.Parent().SpanID().String()
	}
	status := s.Status()
	return dumpedSpan{
		TraceID:       sc.TraceID().String(),
		SpanID:        sc.SpanID().String(),
		ParentSpanID:  parentID,
		Name:          s.Name(),
		Kind:          s.SpanKind().String(),
		StatusCode:    statusCodeString(status.Code),
		StatusMessage: redactURLsInText(status.Description),
		StartTime:     s.StartTime().UTC(),
		EndTime:       s.EndTime().UTC(),
		Attributes:    attributesToMap(s.Attributes()),
		Events:        snapshotEvents(s.Events()),
	}
}
```

#### traceDumpBuffer (type)

```go
type traceDumpBuffer struct {
	spans    []dumpedSpan
	hasError bool
}
```


## ./internal/pruneagent
- package: `pruneagent`
- packageDoc: Package pruneagent investigates local-only / likely-removable branches.
- hasMain: false
- jsonTags: true
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- mechanicalRole: unknown
- mechanicalConfidence: 0.00
- exportedDecls: Card, Evidence, Request, Result, Service, SessionStore, VerdictAskUser, VerdictDrop, VerdictKeep
- exportedFuncs: New, NewWithStore
- exportedMethods: Service.Investigate
- unexportedDecls: cardSystemPrompt, kindPruneInvestigate, sourceLLM, sourceRules
- unexportedFuncs: buildUserPrompt, clampCard, dropCommand, nonEmptyLines, parseCard, synthesizeCard
- unexportedMethods: Service.gather, Service.git, Service.persistLLMFailure, Service.synthesize
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/pruneagent/service.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/pruneagent/store.go

### Exported bodies

#### Request (type)

```go
type Request struct {
	ProjectID     string `json:"project_id"`
	Branch        string `json:"branch"`
	WorktreePath  string `json:"worktree_path"`
	DefaultBranch string `json:"default_branch"`
}
```

#### Evidence (type)

```go
type Evidence struct {
	Branch           string   `json:"branch"`
	WorktreePath     string   `json:"worktree_path"`
	DefaultBranch    string   `json:"default_branch"`
	Dirty            bool     `json:"dirty"`
	RelatedHistories bool     `json:"related_histories"`
	ContentOnDefault bool     `json:"content_on_default"`
	UniqueCommits    []string `json:"unique_commits,omitempty"`
	UniqueCommitN    int      `json:"unique_commit_count"`
	UniqueFiles      []string `json:"unique_files,omitempty"`
	UniqueFileN      int      `json:"unique_file_count"`
	AheadOfDefault   int      `json:"ahead_of_default"`
	BehindDefault    int      `json:"behind_default"`
	Notes            []string `json:"notes,omitempty"`
	GatheredAt       string   `json:"gathered_at"`
}
```

#### Card (type)

```go
type Card struct {
	Verdict  string   `json:"verdict"` // drop | keep | ask_user
	Summary  string   `json:"summary"`
	Bullets  []string `json:"bullets"`
	Command  string   `json:"command,omitempty"`
	Evidence string   `json:"evidence_ref,omitempty"`
}
```

#### Result (type)

```go
type Result struct {
	SessionID string   `json:"session_id"`
	Card      Card     `json:"card"`
	Evidence  Evidence `json:"evidence"`
	Model     string   `json:"model,omitempty"`
	Source    string   `json:"source,omitempty"` // llm | rules
}
```

#### Service (type)

```go
type Service struct {
	Store SessionStore
	Run   cliexec.Exec
	LLM   *llm.Client
}
```

#### New (func)

```go
func New(agentsRoot string, run cliexec.Exec, llmClient *llm.Client) (*Service, error) {
	store, err := agentsession.New(agentsRoot)
	if err != nil {
		return nil, err
	}
	return NewWithStore(store, run, llmClient), nil
}
```

#### NewWithStore (func)

```go
func NewWithStore(store SessionStore, run cliexec.Exec, llmClient *llm.Client) *Service {
	if store == nil {
		panic("pruneagent.NewWithStore: Store is required")
	}
	if run == nil {
		panic("pruneagent.NewWithStore: Exec is required")
	}
	return &Service{Store: store, Run: run, LLM: llmClient}
}
```

#### Service.Investigate (method)

```go
func (s *Service) Investigate(ctx context.Context, req Request) (*Result, error) {
	if s == nil || s.Store == nil {
		return nil, fmt.Errorf("prune agent missing")
	}
	branch := strings.TrimSpace(req.Branch)
	wt := strings.TrimSpace(req.WorktreePath)
	def := strings.TrimSpace(req.DefaultBranch)
	if branch == "" || wt == "" {
		return nil, fmt.Errorf("branch and worktree_path are required")
	}
	if err := localgit.ValidateBranchName(branch); err != nil {
		return nil, fmt.Errorf("branch: %w", err)
	}
	if def == "" {
		def = "main"
	}
	if err := localgit.ValidateBranchName(def); err != nil {
		return nil, fmt.Errorf("default branch: %w", err)
	}
	abs, err := localgit.ExpandPath(wt)
	if err != nil {
		return nil, fmt.Errorf("worktree path: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("worktree path: %w", err)
	}

	meta, err := s.Store.Create(ctx, kindPruneInvestigate, map[string]any{
		"project_id":     strings.TrimSpace(req.ProjectID),
		"branch":         branch,
		"worktree_path":  abs,
		"default_branch": def,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	ev, err := s.gather(ctx, abs, branch, def)
	if err != nil {
		if turnErr := s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{Role: "system", Content: "evidence error: " + err.Error()}); turnErr != nil {
			return nil, fmt.Errorf("evidence: %w (also append turn: %v)", err, turnErr)
		}
		return nil, err
	}
	if err := s.Store.SaveJSON(ctx, meta.ID, agentsession.FileEvidence, ev); err != nil {
		return nil, fmt.Errorf("save evidence: %w", err)
	}

	userPrompt, err := buildUserPrompt(ev)
	if err != nil {
		return nil, err
	}
	if err := s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{Role: "user", Content: userPrompt}); err != nil {
		return nil, fmt.Errorf("append user turn: %w", err)
	}

	card, model, source, err := s.synthesize(ctx, meta.ID, ev, userPrompt)
	if err != nil {
		return nil, err
	}
	card = clampCard(ev, card)

	if err := s.Store.SaveJSON(ctx, meta.ID, agentsession.FileCard, card); err != nil {
		return nil, fmt.Errorf("save card: %w", err)
	}
	if err := s.Store.AppendTurn(ctx, meta.ID, agentsession.Turn{
		Role:    "assistant",
		Content: card.Summary,
		Refs:    map[string]any{"card": true, "verdict": card.Verdict, "source": source, "model": model},
	}); err != nil {
		return nil, fmt.Errorf("append assistant turn: %w", err)
	}

	return &Result{
		SessionID: meta.ID,
		Card:      card,
		Evidence:  ev,
		Model:     model,
		Source:    source,
	}, nil
}
```

#### SessionStore (type)

```go
type SessionStore interface {
	Create(ctx context.Context, kind string, extra map[string]any) (*agentsession.Meta, error)
	SaveJSON(ctx context.Context, id, name string, v any) error
	AppendTurn(ctx context.Context, id string, turn agentsession.Turn) error
	Dir(id string) (string, error)
	Load(ctx context.Context, id string) (*agentsession.Meta, error)
	LoadJSON(ctx context.Context, id, name string, dest any) error
	ReadTurns(ctx context.Context, id string) ([]agentsession.Turn, error)
}
```

### Private one-hop bodies

#### Service.gather (method)

```go
func (s *Service) gather(ctx context.Context, dir, branch, def string) (Evidence, error) {
	ev := Evidence{
		Branch:        branch,
		WorktreePath:  dir,
		DefaultBranch: def,
		GatheredAt:    time.Now().UTC().Format(time.RFC3339),
	}

	if out, err := s.git(ctx, dir, "status", "--porcelain"); err == nil {
		ev.Dirty = strings.TrimSpace(string(out)) != ""
	} else {
		// Fail closed: unknown dirty state must not recommend drop.
		ev.Dirty = true
		ev.Notes = append(ev.Notes, "status: "+err.Error())
	}

	// Are branch and default related?
	if _, err := s.git(ctx, dir, "merge-base", def, branch); err != nil {
		ev.RelatedHistories = false
		ev.Notes = append(ev.Notes, "histories unrelated to "+def)
	} else {
		ev.RelatedHistories = true
	}

	if ev.RelatedHistories {
		if out, err := s.git(ctx, dir, "rev-list", "--left-right", "--count", def+"..."+branch); err == nil {
			var behind, ahead int
			if _, scanErr := fmt.Sscanf(strings.TrimSpace(string(out)), "%d\t%d", &behind, &ahead); scanErr == nil {
				ev.BehindDefault = behind
				ev.AheadOfDefault = ahead
			} else {
				ev.Notes = append(ev.Notes, "rev-list count parse: "+scanErr.Error())
			}
		}
		if out, err := s.git(ctx, dir, "log", "--oneline", def+".."+branch); err == nil {
			lines := nonEmptyLines(string(out))
			ev.UniqueCommitN = len(lines)
			if len(lines) > 20 {
				ev.UniqueCommits = lines[:20]
			} else {
				ev.UniqueCommits = lines
			}
		}
		if out, err := s.git(ctx, dir, "diff", "--name-only", def+"..."+branch); err == nil {
			files := nonEmptyLines(string(out))
			ev.UniqueFileN = len(files)
			if len(files) > 40 {
				ev.UniqueFiles = files[:40]
			} else {
				ev.UniqueFiles = files
			}
		}
	}

	in := localgit.NewInspector(s.Run)
	// Best-effort refresh so ContentOnDefault can prefer origin/<default>.
	if err := in.FetchOrigin(ctx, dir); err != nil {
		ev.Notes = append(ev.Notes, "fetch origin: "+err.Error())
	}
	ok, reason, err := in.ContentOnDefault(ctx, dir, branch, def)
	if err != nil {
		ev.Notes = append(ev.Notes, "content_on_default: "+err.Error())
	} else {
		ev.ContentOnDefault = ok
		if reason != "" {
			ev.Notes = append(ev.Notes, "content_on_default: "+reason)
		}
	}
	return ev, nil
}
```

#### Service.git (method)

```go
func (s *Service) git(ctx context.Context, dir string, args ...string) ([]byte, error) {
	argv := append([]string{"-C", dir}, args...)
	return s.Run.Run(ctx, "git", argv...)
}
```

#### Service.synthesize (method)

```go
func (s *Service) synthesize(ctx context.Context, sessionID string, ev Evidence, userPrompt string) (Card, string, string, error) {
	if s == nil || s.LLM == nil || !s.LLM.Enabled() {
		return synthesizeCard(ev), "", sourceRules, nil
	}
	raw, model, err := s.LLM.Chat(ctx, cardSystemPrompt, userPrompt)
	if err != nil {
		if persistErr := s.persistLLMFailure(ctx, sessionID, map[string]any{
			"error": err.Error(),
			"at":    time.Now().UTC().Format(time.RFC3339),
		}, agentsession.Turn{
			Role:    "system",
			Content: "llm error, falling back to rules: " + err.Error(),
		}); persistErr != nil {
			return Card{}, "", "", fmt.Errorf("llm chat: %w (also persist failure: %v)", err, persistErr)
		}
		return synthesizeCard(ev), "", sourceRules, nil
	}
	card, ok := parseCard(raw)
	if !ok {
		if persistErr := s.persistLLMFailure(ctx, sessionID, map[string]any{
			"error": "unparseable card",
			"raw":   raw,
			"at":    time.Now().UTC().Format(time.RFC3339),
		}, agentsession.Turn{
			Role:    "system",
			Content: "llm response unparseable, falling back to rules",
			Refs:    map[string]any{"raw": raw},
		}); persistErr != nil {
			return Card{}, model, "", fmt.Errorf("persist unparseable card: %w", persistErr)
		}
		return synthesizeCard(ev), model, sourceRules, nil
	}
	return card, model, sourceLLM, nil
}
```

#### buildUserPrompt (func)

```go
func buildUserPrompt(ev Evidence) (string, error) {
	raw, err := json.MarshalIndent(ev, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal evidence: %w", err)
	}
	return "Judge this local branch for deletion. Evidence JSON:\n" + string(raw), nil
}
```

#### clampCard (func)

```go
func clampCard(ev Evidence, card Card) Card {
	switch card.Verdict {
	case VerdictDrop, VerdictKeep, VerdictAskUser:
	default:
		card.Verdict = VerdictAskUser
		if card.Summary == "" {
			card.Summary = "Ambiguous verdict; inspect before deleting."
		}
		card.Command = ""
	}
	if ev.Dirty {
		card.Verdict = VerdictAskUser
		card.Command = ""
		if card.Summary == "" {
			card.Summary = "Working tree is dirty; do not delete until changes are committed or discarded."
		}
	}
	if !ev.RelatedHistories {
		card.Verdict = VerdictAskUser
		card.Command = ""
		if card.Summary == "" {
			card.Summary = "Branch history is unrelated to the default branch; inspect before deleting."
		}
	}
	if ev.ContentOnDefault && card.Verdict != VerdictAskUser {
		card.Verdict = VerdictDrop
		if card.Summary == "" {
			card.Summary = "Branch tip matches default (or is already merged); safe to delete the local branch after switching away."
		}
	}
	if ev.UniqueCommitN > 0 && !ev.ContentOnDefault && card.Verdict == VerdictDrop {
		card.Verdict = VerdictKeep
		card.Command = ""
		if card.Summary == "" {
			card.Summary = "Unique commits remain on this branch; keep until you merge, cherry-pick, or explicitly drop that work."
		}
	}
	if card.Verdict == VerdictDrop {
		card.Command = dropCommand(ev)
	} else {
		card.Command = ""
	}
	return card
}
```


## ./internal/remotegit
- package: `remotegit`
- hasMain: false
- jsonTags: false
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- mechanicalRole: unknown
- mechanicalConfidence: 0.00
- exportedDecls: Client, GitHub, GitLab, HeadsSnapshot, RemoteHead, RepoRef, SummaryOpts, TTLCache
- exportedFuncs: EnrichPruneHints, NewGitHub, NewGitLab, NewTTLCache
- exportedMethods: GitHub.AuthStatus, GitHub.FailedJobs, GitHub.JobLog, GitHub.ListOrgRepos, GitHub.ProjectSummary, GitLab.AuthStatus, GitLab.FailedJobs, GitLab.JobLog, GitLab.ListGroupRepos, GitLab.ProjectSummary, TTLCache.Clear, TTLCache.GetOrLoadHeads, TTLCache.GetOrLoadMerged, TTLCache.SetNow
- unexportedDecls: branchAccum, cacheEntry, reviewInfo, staleAfter, summaryLoader
- unexportedFuncs: applyPruneHint, baseSummary, cacheKey, cloneHeads, cloneMerged, firstNonEmpty, githubHasConflict, gitlabHasConflict, newBranchAccum, newestMergedByBranch, parseTime, pathOrg, projectSummaryShared, resolveDefaultBranch, trimBranch, truncateLog
- unexportedMethods: GitHub.loadCI, GitHub.loadMerged, GitHub.loadOpenReviews, GitHub.seedHeads, GitHub.unauthMsg, GitLab.loadCI, GitLab.loadMerged, GitLab.loadOpenReviews, GitLab.seedHeads, GitLab.unauthMsg, TTLCache.clock, TTLCache.entry, branchAccum.addRemote, branchAccum.ensure, branchAccum.list, branchAccum.setCI, branchAccum.setDefault, branchAccum.setNow, branchAccum.setOpenReview, branchAccum.touchUpdated
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/remotegit/branches.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/remotegit/github.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/remotegit/gitlab.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/remotegit/helpers.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/remotegit/loader.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/remotegit/prune.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/remotegit/ttlcache.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/remotegit/types.go

### Exported bodies

#### GitHub (type)

```go
type GitHub struct {
	Run cliexec.Exec
}
```

#### NewGitHub (func)

```go
func NewGitHub(run cliexec.Exec) *GitHub {
	if run == nil {
		panic("remotegit.NewGitHub: Exec is required")
	}
	return &GitHub{Run: run}
}
```

#### GitHub.AuthStatus (method)

```go
func (g *GitHub) AuthStatus(ctx context.Context) (bool, bool, string) {
	if _, err := g.Run.LookPath("gh"); err != nil {
		return false, false, "gh not on PATH"
	}
	_, err := g.Run.Run(ctx, "gh", "auth", "status")
	if err != nil {
		return true, false, strings.TrimPrefix(err.Error(), "gh auth status: ")
	}
	return true, true, ""
}
```

#### GitHub.ProjectSummary (method)

```go
func (g *GitHub) ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (board.ProjectSummary, error) {
	return projectSummaryShared(ctx, p, opts, g)
}
```

#### GitHub.FailedJobs (method)

```go
func (g *GitHub) FailedJobs(ctx context.Context, p config.Project, runID string) ([]board.FailedJob, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("missing run id")
	}
	raw, err := g.Run.RunJSON(ctx, "gh", "run", "view", runID,
		"--repo", p.Path,
		"--json", "jobs",
	)
	if err != nil {
		return nil, fmt.Errorf("github failed jobs: %w", err)
	}
	var payload struct {
		Jobs []struct {
			DatabaseID int64  `json:"databaseId"`
			Name       string `json:"name"`
			Conclusion string `json:"conclusion"`
			URL        string `json:"url"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("parse github jobs: %w", err)
	}
	var out []board.FailedJob
	for _, j := range payload.Jobs {
		if j.Conclusion != "failure" {
			continue
		}
		out = append(out, board.FailedJob{
			ID:     fmt.Sprintf("github:%d", j.DatabaseID),
			Name:   j.Name,
			WebURL: j.URL,
		})
	}
	return out, nil
}
```

#### GitHub.JobLog (method)

```go
func (g *GitHub) JobLog(ctx context.Context, p config.Project, runID, jobID string) (string, error) {
	runID = strings.TrimSpace(runID)
	jobID = strings.TrimPrefix(strings.TrimSpace(jobID), "github:")
	if runID == "" {
		return "", fmt.Errorf("missing run id")
	}
	args := []string{"run", "view", runID, "--repo", p.Path}
	if jobID != "" {
		args = append(args, "--job", jobID, "--log")
	} else {
		args = append(args, "--log-failed")
	}
	out, err := g.Run.Run(ctx, "gh", args...)
	if err != nil {
		return "", fmt.Errorf("github job log: %w", err)
	}
	return truncateLog(string(out)), nil
}
```

#### GitHub.ListOrgRepos (method)

```go
func (g *GitHub) ListOrgRepos(ctx context.Context, org string) ([]RepoRef, error) {
	org = strings.TrimSpace(org)
	if org == "" {
		return nil, fmt.Errorf("missing github org")
	}
	raw, err := g.Run.RunJSON(ctx, "gh", "repo", "list", org,
		"--limit", "1000",
		"--no-archived",
		"--json", "nameWithOwner,name,isArchived",
	)
	if err != nil {
		return nil, fmt.Errorf("github list repos: %w", err)
	}
	var rows []struct {
		NameWithOwner string `json:"nameWithOwner"`
		Name          string `json:"name"`
		IsArchived    bool   `json:"isArchived"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse gh repo list: %w", err)
	}
	out := make([]RepoRef, 0, len(rows))
	for _, r := range rows {
		if r.IsArchived {
			continue
		}
		path := strings.Trim(r.NameWithOwner, "/")
		name := r.Name
		if name == "" {
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}
		out = append(out, RepoRef{
			Host: config.HostGitHub,
			Path: path,
			Name: name,
		})
	}
	return out, nil
}
```

#### GitLab (type)

```go
type GitLab struct {
	Run cliexec.Exec
}
```

#### NewGitLab (func)

```go
func NewGitLab(run cliexec.Exec) *GitLab {
	if run == nil {
		panic("remotegit.NewGitLab: Exec is required")
	}
	return &GitLab{Run: run}
}
```

#### GitLab.AuthStatus (method)

```go
func (g *GitLab) AuthStatus(ctx context.Context) (bool, bool, string) {
	if _, err := g.Run.LookPath("glab"); err != nil {
		return false, false, "glab not on PATH (brew install glab)"
	}
	_, err := g.Run.Run(ctx, "glab", "auth", "status")
	if err != nil {
		return true, false, strings.TrimPrefix(err.Error(), "glab auth status: ")
	}
	return true, true, ""
}
```

#### GitLab.ProjectSummary (method)

```go
func (g *GitLab) ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (board.ProjectSummary, error) {
	return projectSummaryShared(ctx, p, opts, g)
}
```

#### GitLab.FailedJobs (method)

```go
func (g *GitLab) FailedJobs(ctx context.Context, p config.Project, pipelineID string) ([]board.FailedJob, error) {
	if strings.TrimSpace(pipelineID) == "" {
		return nil, fmt.Errorf("missing pipeline id")
	}
	raw, err := g.Run.RunJSON(ctx, "glab", "ci", "view", pipelineID,
		"-R", p.Path,
		"--output", "json",
	)
	if err != nil {
		return nil, fmt.Errorf("gitlab failed jobs: %w", err)
	}
	var payload struct {
		Jobs []struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Stage  string `json:"stage"`
			Status string `json:"status"`
			WebURL string `json:"web_url"`
		} `json:"jobs"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("parse gitlab jobs: %w", err)
	}
	var out []board.FailedJob
	for _, j := range payload.Jobs {
		if j.Status != "failed" {
			continue
		}
		out = append(out, board.FailedJob{
			ID:     fmt.Sprintf("gitlab:%d", j.ID),
			Name:   j.Name,
			Stage:  j.Stage,
			WebURL: j.WebURL,
		})
	}
	return out, nil
}
```

#### GitLab.JobLog (method)

```go
func (g *GitLab) JobLog(ctx context.Context, p config.Project, _, jobID string) (string, error) {
	jobID = strings.TrimPrefix(strings.TrimSpace(jobID), "gitlab:")
	if jobID == "" {
		return "", fmt.Errorf("missing job id")
	}
	out, err := g.Run.Run(ctx, "glab", "ci", "trace", jobID, "-R", p.Path)
	if err != nil {
		return "", fmt.Errorf("gitlab job log: %w", err)
	}
	return truncateLog(string(out)), nil
}
```

#### GitLab.ListGroupRepos (method)

```go
func (g *GitLab) ListGroupRepos(ctx context.Context, group string) ([]RepoRef, error) {
	group = strings.TrimSpace(group)
	if group == "" {
		return nil, fmt.Errorf("missing gitlab group")
	}
	var out []RepoRef
	for page := 1; page <= 50; page++ {
		raw, err := g.Run.RunJSON(ctx, "glab", "repo", "list",
			"--group", group,
			"--include-subgroups",
			"--archived=false",
			"--per-page", "100",
			"--page", fmt.Sprintf("%d", page),
			"--output", "json",
		)
		if err != nil {
			return nil, fmt.Errorf("gitlab list repos: %w", err)
		}
		var rows []struct {
			PathWithNamespace string `json:"path_with_namespace"`
			Name              string `json:"name"`
			Archived          bool   `json:"archived"`
		}
		if err := json.Unmarshal(raw, &rows); err != nil {
			return nil, fmt.Errorf("parse glab repo list: %w", err)
		}
		if len(rows) == 0 {
			break
		}
		for _, r := range rows {
			if r.Archived {
				continue
			}
			path := strings.Trim(r.PathWithNamespace, "/")
			parts := strings.Split(path, "/")
			if len(parts) < 2 {
				continue
			}
			name := r.Name
			if name == "" {
				name = parts[len(parts)-1]
			}
			out = append(out, RepoRef{
				Host: config.HostGitLab,
				Path: path,
				Name: name,
			})
		}
		if len(rows) < 100 {
			break
		}
	}
	return out, nil
}
```

#### EnrichPruneHints (func)

```go
func EnrichPruneHints(summary *board.ProjectSummary) {
	if summary == nil || summary.Local == nil {
		return
	}
	if !summary.RemoteNamesOK {
		return
	}

	remote := make(map[string]struct{})
	open := make(map[string]struct{})
	for _, name := range summary.RemoteNames {
		name = trimBranch(name)
		if name != "" {
			remote[name] = struct{}{}
		}
	}
	for _, b := range summary.Branches {
		name := trimBranch(b.Name)
		if name == "" {
			continue
		}
		// Fail closed: a branch still shown as a remote row must not be prune-hinted,
		// even if RemoteNames briefly lagged behind the Branches payload.
		remote[name] = struct{}{}
		if b.OpenReview {
			open[name] = struct{}{}
		}
	}

	defaultName := resolveDefaultBranch(summary)
	if defaultName != "" {
		summary.Local.DefaultBranch = defaultName
	}

	mergedByBranch := newestMergedByBranch(summary.Merged)
	for i := range summary.Local.Worktrees {
		applyPruneHint(&summary.Local.Worktrees[i], defaultName, remote, open, mergedByBranch, summary.MergedOK)
	}
}
```

#### TTLCache (type)

```go
type TTLCache struct {
	mu    sync.Mutex
	now   func() time.Time
	byKey map[string]*cacheEntry
}
```

#### NewTTLCache (func)

```go
func NewTTLCache() *TTLCache {
	return &TTLCache{
		now:   time.Now,
		byKey: map[string]*cacheEntry{},
	}
}
```

#### TTLCache.SetNow (method)

```go
func (c *TTLCache) SetNow(now func() time.Time) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if now == nil {
		c.now = time.Now
		return
	}
	c.now = now
}
```

#### TTLCache.Clear (method)

```go
func (c *TTLCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byKey = map[string]*cacheEntry{}
}
```

#### TTLCache.GetOrLoadHeads (method)

```go
func (c *TTLCache) GetOrLoadHeads(key string, ttl time.Duration, fresh bool, load func() (HeadsSnapshot, error)) (HeadsSnapshot, error) {
	if c == nil || ttl <= 0 || fresh {
		return load()
	}
	c.mu.Lock()
	e := c.entry(key)
	if e.hasHeads && c.clock().Sub(e.headsAt) < ttl {
		out := cloneHeads(e.heads)
		c.mu.Unlock()
		return out, nil
	}
	c.mu.Unlock()

	snap, err := load()
	if err != nil {
		return HeadsSnapshot{}, err
	}

	c.mu.Lock()
	e = c.entry(key)
	e.heads = cloneHeads(snap)
	e.headsAt = c.clock()
	e.hasHeads = true
	c.mu.Unlock()
	return snap, nil
}
```

#### TTLCache.GetOrLoadMerged (method)

```go
func (c *TTLCache) GetOrLoadMerged(key string, ttl time.Duration, fresh bool, load func() ([]board.MergedReview, error)) (merged []board.MergedReview, ok bool) {
	if c == nil || ttl <= 0 || fresh {
		list, err := load()
		if err != nil {
			return nil, false
		}
		return cloneMerged(list), true
	}
	c.mu.Lock()
	e := c.entry(key)
	if e.hasMerged && e.mergedOK && c.clock().Sub(e.mergedAt) < ttl {
		out := cloneMerged(e.merged)
		c.mu.Unlock()
		return out, true
	}
	c.mu.Unlock()

	list, err := load()
	if err != nil {
		// Fail closed for prune: expired cache must not keep MergedOK after a failed refresh.
		return nil, false
	}

	c.mu.Lock()
	e = c.entry(key)
	e.merged = cloneMerged(list)
	e.mergedAt = c.clock()
	e.hasMerged = true
	e.mergedOK = true
	c.mu.Unlock()
	return cloneMerged(list), true
}
```

#### HeadsSnapshot (type)

```go
type HeadsSnapshot struct {
	DefaultBranch string
	Heads         []RemoteHead
}
```

#### RemoteHead (type)

```go
type RemoteHead struct {
	Name      string
	UpdatedAt string
	WebURL    string
}
```

#### SummaryOpts (type)

```go
type SummaryOpts struct {
	Fresh     bool
	Cache     *TTLCache
	HeadsTTL  time.Duration // 0 = always refetch heads
	MergedTTL time.Duration // 0 = always refetch merged
}
```

#### Client (type)

```go
type Client interface {
	AuthStatus(ctx context.Context) (installed, authed bool, detail string)
	ProjectSummary(ctx context.Context, p config.Project, opts SummaryOpts) (board.ProjectSummary, error)
	FailedJobs(ctx context.Context, p config.Project, runID string) ([]board.FailedJob, error)
	JobLog(ctx context.Context, p config.Project, runID, jobID string) (string, error)
}
```

#### RepoRef (type)

```go
type RepoRef struct {
	Host     config.Host
	Path     string // owner/repo
	Name     string
	Archived bool
}
```

### Private one-hop bodies

#### TTLCache.clock (method)

```go
func (c *TTLCache) clock() time.Time {
	if c == nil || c.now == nil {
		return time.Now()
	}
	return c.now()
}
```

#### TTLCache.entry (method)

```go
func (c *TTLCache) entry(key string) *cacheEntry {
	if c.byKey == nil {
		c.byKey = map[string]*cacheEntry{}
	}
	e, ok := c.byKey[key]
	if !ok {
		e = &cacheEntry{}
		c.byKey[key] = e
	}
	return e
}
```

#### applyPruneHint (func)

```go
func applyPruneHint(
	wt *board.LocalWorktree,
	defaultName string,
	remote, open map[string]struct{},
	mergedByBranch map[string]board.MergedReview,
	mergedOK bool,
) {
	if wt == nil {
		return
	}
	branch := trimBranch(wt.Branch)
	if branch == "" || wt.Detached || wt.Bare {
		return
	}
	if defaultName != "" && branch == defaultName {
		return
	}
	if _, onRemote := remote[branch]; onRemote {
		return
	}
	if wt.Dirty {
		return
	}
	// Open review blocks every prune hint (including content-on-default / merged).
	if _, stillOpen := open[branch]; stillOpen {
		return
	}
	if m, ok := mergedByBranch[branch]; ok {
		wt.PruneHint = board.PruneSafe
		wt.MergedID = m.ID
		wt.MergedURL = m.URL
		wt.MergedAt = m.MergedAt
		return
	}
	if wt.ContentOnDefault {
		wt.PruneHint = board.PruneSafe
		return
	}
	if !mergedOK {
		return
	}
	wt.PruneHint = board.PruneLikely
}
```

#### branchAccum.list (method)

```go
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
	const max = 20
	if len(out) > max {
		out = out[:max]
	}
	return out
}
```

#### cacheEntry (type)

```go
type cacheEntry struct {
	heads     HeadsSnapshot
	headsAt   time.Time
	hasHeads  bool
	merged    []board.MergedReview
	mergedOK  bool
	mergedAt  time.Time
	hasMerged bool
}
```

#### cloneHeads (func)

```go
func cloneHeads(in HeadsSnapshot) HeadsSnapshot {
	out := HeadsSnapshot{DefaultBranch: in.DefaultBranch}
	if len(in.Heads) > 0 {
		out.Heads = append([]RemoteHead(nil), in.Heads...)
	}
	return out
}
```

#### cloneMerged (func)

```go
func cloneMerged(in []board.MergedReview) []board.MergedReview {
	if len(in) == 0 {
		return nil
	}
	return append([]board.MergedReview(nil), in...)
}
```

#### newestMergedByBranch (func)

```go
func newestMergedByBranch(merged []board.MergedReview) map[string]board.MergedReview {
	out := make(map[string]board.MergedReview, len(merged))
	for _, m := range merged {
		name := trimBranch(m.Branch)
		if name == "" {
			continue
		}
		prev, ok := out[name]
		if !ok {
			out[name] = m
			continue
		}
		tNew, okNew := parseTime(m.MergedAt)
		tOld, okOld := parseTime(prev.MergedAt)
		if okNew && (!okOld || tNew.After(tOld)) {
			out[name] = m
		}
	}
	return out
}
```

#### projectSummaryShared (func)

```go
func projectSummaryShared(ctx context.Context, p config.Project, opts SummaryOpts, loader summaryLoader) (board.ProjectSummary, error) {
	summary := baseSummary(p)

	// Step 1: auth.
	installed, authed, detail := loader.AuthStatus(ctx)
	if !installed {
		summary.Error = detail
		return summary, nil
	}
	if !authed {
		summary.Error = loader.unauthMsg(detail)
		return summary, nil
	}

	repo := p.Path
	key := cacheKey(string(p.Host), repo)
	branches := newBranchAccum()

	// Step 2: seed heads (default branch + first page only).
	heads, headsErr := opts.Cache.GetOrLoadHeads(key, opts.HeadsTTL, opts.Fresh, func() (HeadsSnapshot, error) {
		return loader.seedHeads(ctx, repo)
	})
	if headsErr != nil {
		if summary.Error == "" {
			summary.Error = headsErr.Error()
		}
		summary.RemoteNamesOK = false
	} else {
		summary.RemoteNamesOK = true
		if heads.DefaultBranch != "" {
			branches.setDefault(heads.DefaultBranch)
		}
		names := make([]string, 0, len(heads.Heads))
		for _, h := range heads.Heads {
			branches.addRemote(h.Name, h.UpdatedAt, h.WebURL)
			if n := trimBranch(h.Name); n != "" {
				names = append(names, n)
			}
		}
		summary.RemoteNames = names
	}

	// Step 3: open reviews (seeds PR/MR branches into branchAccum).
	if err := loader.loadOpenReviews(ctx, repo, &summary, branches); err != nil && summary.Error == "" {
		summary.Error = err.Error()
	}

	// Step 4: CI (attaches to branches seeded by heads or reviews).
	if err := loader.loadCI(ctx, repo, &summary, branches); err != nil && summary.Error == "" {
		summary.Error = err.Error()
	}

	// Step 5: merged reviews (for prune hints).
	merged, mergedOK := opts.Cache.GetOrLoadMerged(key, opts.MergedTTL, opts.Fresh, func() ([]board.MergedReview, error) {
		return loader.loadMerged(ctx, repo)
	})
	summary.Merged = merged
	summary.MergedOK = mergedOK
	summary.Branches = branches.list()
	return summary, nil
}
```

#### resolveDefaultBranch (func)

```go
func resolveDefaultBranch(summary *board.ProjectSummary) string {
	if summary == nil {
		return ""
	}
	for _, b := range summary.Branches {
		if !b.Default {
			continue
		}
		if name := trimBranch(b.Name); name != "" {
			return name
		}
	}
	if summary.Local == nil {
		return ""
	}
	return strings.TrimSpace(summary.Local.DefaultBranch)
}
```

#### trimBranch (func)

```go
func trimBranch(name string) string {
	return strings.TrimSpace(name)
}
```

#### truncateLog (func)

```go
func truncateLog(s string) string {
	const max = 96 * 1024
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}
```


## ./internal/server
- package: `server`
- hasMain: false
- jsonTags: false
- goEmbed: true
- embedsStatic: true
- importsNetHTTP: true
- importsOsExec: false
- importsGrpc: false
- importsOtel: true
- importsPrometheus: false
- deliveryHint: server-ui
- mechanicalRole: server
- mechanicalConfidence: 0.90
- mechanicalEvidence: delivery:http, go_embed, delivery:ui, embeds_static, imports_net_http, http_surface_ident
- exportedDecls: Options
- exportedFuncs: NewMux
- exportedMethods: (none)
- unexportedDecls: configLive, errDashboardUnavailable, errForgeClientMissing, errMethodNotAllowed, errUnknownProject, maxJSONBody, staticFS
- unexportedFuncs: decodeJSONBody, newConfigLive, projectFingerprint, projectsChanged, writeJSON
- unexportedMethods: configLive.snapshot
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/server/configlive.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/server/server.go

### Exported bodies

#### Options (type)

```go
type Options struct {
	Addr string
	// ConfigPath, when set, reloads the full config file when mtime advances
	// (so gitboard sync and ui edits update a running board).
	ConfigPath string
	// Doc is the initial full config snapshot (from config.Load).
	Doc         config.File
	Dash        *dashboard.Service
	Commands    *dashboard.Commands
	Triage      *triage.Analyzer
	Prune       *pruneagent.Service
	PollSeconds int
}
```

#### NewMux (func)

```go
func NewMux(opts Options) http.Handler {
	mux := http.NewServeMux()
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("server: embed static: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	clearCaches := func() {
		if opts.Dash != nil {
			opts.Dash.ClearCaches()
		}
	}
	poll := opts.PollSeconds
	if poll == 0 {
		poll = opts.Doc.EffectivePollSeconds()
	}
	live := newConfigLive(opts.ConfigPath, opts.Doc, poll, clearCaches)

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte("ok\n"))
		}
	})

	mux.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		_, poll := live.snapshot()
		writeJSON(w, map[string]any{
			"poll_interval_seconds": poll,
		})
	})

	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Dash == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		fresh := r.URL.Query().Get("fresh") == "1" || strings.EqualFold(r.URL.Query().Get("fresh"), "true")
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		doc, poll := live.snapshot()
		payload := opts.Dash.Collect(ctx, doc, fresh)
		payload.PollIntervalSeconds = poll
		if err := ctx.Err(); err != nil {
			http.Error(w, "dashboard timed out or canceled", http.StatusGatewayTimeout)
			return
		}
		writeJSON(w, payload)
	})

	mux.HandleFunc("/api/failures", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Dash == nil {
			http.Error(w, errDashboardUnavailable, http.StatusServiceUnavailable)
			return
		}
		doc, _ := live.snapshot()
		projectID := strings.TrimSpace(r.URL.Query().Get("project"))
		runID := strings.TrimSpace(r.URL.Query().Get("run_id"))
		p, ok := dashboard.FindProject(doc.Projects, projectID)
		if !ok {
			http.Error(w, errUnknownProject, http.StatusBadRequest)
			return
		}
		client := dashboard.ClientFor(opts.Dash, p)
		if client == nil {
			http.Error(w, errForgeClientMissing, http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
		defer cancel()
		jobs, err := client.FailedJobs(ctx, p, runID)
		if err != nil {
			http.Error(w, "failed to list jobs", http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]any{"jobs": jobs})
	})

	mux.HandleFunc("/api/triage", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}
		if opts.Dash == nil || opts.Triage == nil {
			http.Error(w, "triage unavailable", http.StatusServiceUnavailable)
			return
		}
		var body triage.Request
		if err := decodeJSONBody(w, r, &body); err != nil {
			return
		}
		doc, _ := live.snapshot()
		p, ok := dashboard.FindProject(doc.Projects, body.ProjectID)
		if !ok {
			http.Error(w, errUnknownProject, http.StatusBadRequest)
			return
		}
		client := dashboard.ClientFor(opts.Dash, p)
		if client == nil {
			http.Error(w, errForgeClientMissing, http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
		defer cancel()
		if strings.TrimSpace(body.Log) == "" {
			logText, err := client.JobLog(ctx, p, body.RunID, body.JobID)
			if err != nil {
				http.Error(w, "failed to fetch job log", http.StatusBadGateway)
				return
			}
			body.Log = logText
		}
		resp, err := opts.Tria
// ... truncated
```

### Private one-hop bodies

#### configLive.snapshot (method)

```go
func (c *configLive) snapshot() (config.File, int) {
	if c == nil {
		return config.File{}, config.DefaultPollSeconds
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.path == "" {
		return c.doc, c.poll
	}
	st, err := os.Stat(c.path)
	if err != nil {
		return c.doc, c.poll
	}
	mod := st.ModTime()
	if !mod.After(c.mod) {
		return c.doc, c.poll
	}
	doc, err := config.Load(c.path)
	if err != nil {
		log.Printf("gitboard: config reload failed path=%s: %v", c.path, err)
		return c.doc, c.poll
	}
	if projectsChanged(c.doc.Projects, doc.Projects) {
		if c.clearCaches != nil {
			c.clearCaches()
		}
	}
	c.doc = doc
	c.mod = mod
	c.poll = doc.EffectivePollSeconds()
	return c.doc, c.poll
}
```

#### decodeJSONBody (func)

```go
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	defer func() { _ = r.Body.Close() }()
	limited := io.LimitReader(r.Body, maxJSONBody+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return err
	}
	if len(raw) > maxJSONBody {
		http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
		return errors.New("body too large")
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return err
	}
	return nil
}
```

#### newConfigLive (func)

```go
func newConfigLive(path string, doc config.File, poll int, clearCaches func()) *configLive {
	cl := &configLive{
		path:        path,
		doc:         doc,
		poll:        poll,
		clearCaches: clearCaches,
	}
	if path != "" {
		if st, err := os.Stat(path); err == nil {
			cl.mod = st.ModTime()
		}
	}
	return cl
}
```

#### writeJSON (func)

```go
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("gitboard: write json: %v", err)
	}
}
```


## ./internal/syncproj
- package: `syncproj`
- hasMain: false
- jsonTags: false
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- mechanicalRole: unknown
- mechanicalConfidence: 0.00
- exportedDecls: Candidate, ForgeLister, Lister
- exportedFuncs: AddProject, ApplySelection, Discover, FormatCandidates, ParseSelection, ProjectFromRef, PromptSyncSources, RemoveProject, SelectInteractive
- exportedMethods: ForgeLister.ListGitHub, ForgeLister.ListGitLab
- unexportedDecls: (none)
- unexportedFuncs: candidateKey, slugID, splitCSV, trackKey, uniqueID
- unexportedMethods: (none)
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/syncproj/select.go, ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/syncproj/sync.go

### Exported bodies

#### SelectInteractive (func)

```go
func SelectInteractive(cands []Candidate) ([]int, error) {
	if len(cands) == 0 {
		return nil, fmt.Errorf("no candidates")
	}
	byKey := map[string]int{}
	opts := make([]huh.Option[string], 0, len(cands))
	var selected []string
	for _, c := range cands {
		key := candidateKey(c)
		byKey[key] = c.Index
		label := fmt.Sprintf("%s  %s", c.Host, c.Path)
		opts = append(opts, huh.NewOption(label, key))
		if c.Tracked {
			selected = append(selected, key)
		}
	}
	height := len(cands) + 2
	if height > 18 {
		height = 18
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Track repositories").
				Description("↑↓ move · space toggle · / filter · enter confirm").
				Options(opts...).
				Filterable(true).
				Height(height).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}
	out := make([]int, 0, len(selected))
	seen := map[int]struct{}{}
	for _, key := range selected {
		idx, ok := byKey[key]
		if !ok {
			continue
		}
		if _, dup := seen[idx]; dup {
			continue
		}
		seen[idx] = struct{}{}
		out = append(out, idx)
	}
	return out, nil
}
```

#### PromptSyncSources (func)

```go
func PromptSyncSources() (orgs, groups []string, err error) {
	var orgLine, groupLine string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("GitHub organizations").
				Description("Comma-separated (empty to skip)").
				Placeholder("behaviorengineering, xynova").
				Value(&orgLine),
			huh.NewInput().
				Title("GitLab groups").
				Description("Comma-separated (empty to skip)").
				Placeholder("behaviorengineering").
				Value(&groupLine),
		),
	)
	if err := form.Run(); err != nil {
		return nil, nil, err
	}
	return splitCSV(orgLine), splitCSV(groupLine), nil
}
```

#### Lister (type)

```go
type Lister interface {
	ListGitHub(ctx context.Context, org string) ([]remotegit.RepoRef, error)
	ListGitLab(ctx context.Context, group string) ([]remotegit.RepoRef, error)
}
```

#### ForgeLister (type)

```go
type ForgeLister struct {
	GitHub *remotegit.GitHub
	GitLab *remotegit.GitLab
}
```

#### ForgeLister.ListGitHub (method)

```go
func (f ForgeLister) ListGitHub(ctx context.Context, org string) ([]remotegit.RepoRef, error) {
	if f.GitHub == nil {
		return nil, fmt.Errorf("github client missing")
	}
	return f.GitHub.ListOrgRepos(ctx, org)
}
```

#### ForgeLister.ListGitLab (method)

```go
func (f ForgeLister) ListGitLab(ctx context.Context, group string) ([]remotegit.RepoRef, error) {
	if f.GitLab == nil {
		return nil, fmt.Errorf("gitlab client missing")
	}
	return f.GitLab.ListGroupRepos(ctx, group)
}
```

#### Candidate (type)

```go
type Candidate struct {
	Host    config.Host
	Path    string
	Name    string
	Tracked bool
	Index   int // 1-based display index
}
```

#### Discover (func)

```go
func Discover(ctx context.Context, lister Lister, doc config.File, hostFilter string) ([]Candidate, error) {
	if lister == nil {
		return nil, fmt.Errorf("lister missing")
	}
	hostFilter = strings.ToLower(strings.TrimSpace(hostFilter))
	tracked := map[string]struct{}{}
	for _, p := range doc.Projects {
		tracked[trackKey(p.Host, p.Path)] = struct{}{}
	}

	seen := map[string]struct{}{}
	var refs []remotegit.RepoRef

	if hostFilter == "" || hostFilter == "github" {
		for _, org := range doc.Sync.GitHub.Orgs {
			list, err := lister.ListGitHub(ctx, org)
			if err != nil {
				return nil, fmt.Errorf("github org %s: %w", org, err)
			}
			refs = append(refs, list...)
		}
	}
	if hostFilter == "" || hostFilter == "gitlab" {
		for _, group := range doc.Sync.GitLab.Groups {
			list, err := lister.ListGitLab(ctx, group)
			if err != nil {
				return nil, fmt.Errorf("gitlab group %s: %w", group, err)
			}
			refs = append(refs, list...)
		}
	}

	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Host != refs[j].Host {
			return refs[i].Host < refs[j].Host
		}
		return refs[i].Path < refs[j].Path
	})

	var out []Candidate
	for _, r := range refs {
		key := trackKey(r.Host, r.Path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		_, isTracked := tracked[key]
		out = append(out, Candidate{
			Host:    r.Host,
			Path:    r.Path,
			Name:    r.Name,
			Tracked: isTracked,
		})
	}
	for i := range out {
		out[i].Index = i + 1
	}
	return out, nil
}
```

#### ProjectFromRef (func)

```go
func ProjectFromRef(host config.Host, path, name string) config.Project {
	path = strings.Trim(path, "/")
	id := slugID(name, path)
	label := name
	if label == "" {
		parts := strings.Split(path, "/")
		label = parts[len(parts)-1]
	}
	return config.Project{
		ID:    id,
		Label: label,
		Host:  host,
		Path:  path,
	}
}
```

#### ApplySelection (func)

```go
func ApplySelection(cands []Candidate, selected []int, existing []config.Project) ([]config.Project, error) {
	byIndex := map[int]Candidate{}
	for _, c := range cands {
		byIndex[c.Index] = c
	}
	keepLocal := map[string]string{}
	for _, p := range existing {
		if strings.TrimSpace(p.LocalPath) == "" {
			continue
		}
		keepLocal[trackKey(p.Host, p.Path)] = p.LocalPath
	}
	var projects []config.Project
	usedIDs := map[string]struct{}{}
	for _, n := range selected {
		c, ok := byIndex[n]
		if !ok {
			return nil, fmt.Errorf("unknown selection index %d", n)
		}
		p := ProjectFromRef(c.Host, c.Path, c.Name)
		if lp, ok := keepLocal[trackKey(p.Host, p.Path)]; ok {
			p.LocalPath = lp
		}
		p.ID = uniqueID(p.ID, usedIDs)
		usedIDs[p.ID] = struct{}{}
		projects = append(projects, p)
	}
	return projects, nil
}
```

#### AddProject (func)

```go
func AddProject(projects []config.Project, host config.Host, path string) ([]config.Project, error) {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("path must be owner/repo")
	}
	name := parts[len(parts)-1]
	p := ProjectFromRef(host, path, name)
	key := trackKey(host, path)
	for i, existing := range projects {
		if trackKey(existing.Host, existing.Path) == key {
			p.ID = existing.ID
			p.LocalPath = existing.LocalPath
			projects[i] = p
			return projects, nil
		}
	}
	used := map[string]struct{}{}
	for _, existing := range projects {
		used[existing.ID] = struct{}{}
	}
	p.ID = uniqueID(p.ID, used)
	return append(projects, p), nil
}
```

#### RemoveProject (func)

```go
func RemoveProject(projects []config.Project, id string) ([]config.Project, bool) {
	id = strings.TrimSpace(id)
	var out []config.Project
	found := false
	for _, p := range projects {
		if p.ID == id {
			found = true
			continue
		}
		out = append(out, p)
	}
	return out, found
}
```

#### ParseSelection (func)

```go
func ParseSelection(line string, cands []Candidate) ([]int, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		var keep []int
		for _, c := range cands {
			if c.Tracked {
				keep = append(keep, c.Index)
			}
		}
		return keep, nil
	}
	if strings.EqualFold(line, "all") {
		out := make([]int, len(cands))
		for i, c := range cands {
			out[i] = c.Index
		}
		return out, nil
	}
	parts := strings.FieldsFunc(line, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	var out []int
	seen := map[int]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("bad selection %q", part)
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out, nil
}
```

#### FormatCandidates (func)

```go
func FormatCandidates(w io.Writer, cands []Candidate) {
	for _, c := range cands {
		mark := " "
		if c.Tracked {
			mark = "x"
		}
		if _, err := fmt.Fprintf(w, "[%d] [%s] %s %s\n", c.Index, mark, c.Host, c.Path); err != nil {
			return
		}
	}
}
```

### Private one-hop bodies

#### candidateKey (func)

```go
func candidateKey(c Candidate) string {
	return string(c.Host) + "\x00" + strings.ToLower(strings.Trim(c.Path, "/"))
}
```

#### slugID (func)

```go
func slugID(name, path string) string {
	base := name
	if base == "" {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		base = parts[len(parts)-1]
	}
	base = strings.ToLower(base)
	var b strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" {
		id = "repo"
	}
	return id
}
```

#### splitCSV (func)

```go
func splitCSV(line string) []string {
	parts := strings.FieldsFunc(line, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
```

#### trackKey (func)

```go
func trackKey(host config.Host, path string) string {
	return string(host) + ":" + strings.ToLower(strings.Trim(path, "/"))
}
```

#### uniqueID (func)

```go
func uniqueID(base string, used map[string]struct{}) string {
	if _, ok := used[base]; !ok {
		return base
	}
	for n := 2; ; n++ {
		id := fmt.Sprintf("%s-%d", base, n)
		if _, ok := used[id]; !ok {
			return id
		}
	}
}
```


## ./internal/triage
- package: `triage`
- hasMain: false
- jsonTags: true
- goEmbed: false
- embedsStatic: false
- importsNetHTTP: false
- importsOsExec: false
- importsGrpc: false
- importsOtel: false
- importsPrometheus: false
- mechanicalRole: unknown
- mechanicalConfidence: 0.00
- exportedDecls: Analyzer, Request, Response
- exportedFuncs: New
- exportedMethods: Analyzer.Analyze, Analyzer.Enabled
- unexportedDecls: (none)
- unexportedFuncs: buildPrompt, parseAnswer
- unexportedMethods: (none)
- errorTypes: (none)
- files: ../../../../../../../var/folders/6h/b51xqt3568q2p96564n8qzh40000gn/T/majordomo-typology-1600230215/internal/triage/triage.go

### Exported bodies

#### Request (type)

```go
type Request struct {
	ProjectID string `json:"project_id"`
	RunID     string `json:"run_id,omitempty"`
	JobID     string `json:"job_id,omitempty"`
	JobName   string `json:"job_name,omitempty"`
	Log       string `json:"log,omitempty"`
}
```

#### Response (type)

```go
type Response struct {
	Summary     string   `json:"summary"`
	RootCause   string   `json:"root_cause"`
	FixSteps    []string `json:"fix_steps"`
	Confidence  string   `json:"confidence"`
	RawAnswer   string   `json:"raw_answer,omitempty"`
	Model       string   `json:"model,omitempty"`
	Unavailable string   `json:"unavailable,omitempty"`
}
```

#### Analyzer (type)

```go
type Analyzer struct {
	LLM *llm.Client
}
```

#### New (func)

```go
func New(cfg config.LLM) *Analyzer {
	return &Analyzer{LLM: llm.New(cfg)}
}
```

#### Analyzer.Enabled (method)

```go
func (a *Analyzer) Enabled() bool {
	return a != nil && a.LLM.Enabled()
}
```

#### Analyzer.Analyze (method)

```go
func (a *Analyzer) Analyze(ctx context.Context, req Request) (Response, error) {
	if !a.Enabled() {
		return Response{
			Unavailable: "set llm.base_url in config.yaml (or GITBOARD_LLM_BASE_URL / POLYPUS_BASE_URL) for AI triage",
		}, nil
	}
	logText := strings.TrimSpace(req.Log)
	if logText == "" {
		return Response{}, fmt.Errorf("empty log excerpt")
	}
	prompt := buildPrompt(req, logText)
	raw, model, err := a.LLM.Chat(ctx,
		"You triage CI failures for a Go monorepo. Do not invent file paths or line numbers absent from the log.",
		prompt,
	)
	if err != nil {
		return Response{}, fmt.Errorf("triage chat: %w", err)
	}
	return parseAnswer(raw, model), nil
}
```

### Private one-hop bodies

#### buildPrompt (func)

```go
func buildPrompt(req Request, logText string) string {
	var b strings.Builder
	b.WriteString("Analyze this CI failure. Use ONLY the log excerpt below.\n")
	if req.JobName != "" {
		b.WriteString("Job: ")
		b.WriteString(req.JobName)
		b.WriteByte('\n')
	}
	b.WriteString("\nRespond in plain text with sections:\n")
	b.WriteString("SUMMARY: one sentence\n")
	b.WriteString("ROOT_CAUSE: short paragraph\n")
	b.WriteString("FIX_STEPS: numbered list\n")
	b.WriteString("CONFIDENCE: high|medium|low\n\n")
	b.WriteString("LOG:\n")
	b.WriteString(logText)
	return b.String()
}
```

#### parseAnswer (func)

```go
func parseAnswer(raw, model string) Response {
	out := Response{RawAnswer: raw, Model: model}
	lines := strings.Split(raw, "\n")
	var section string
	var steps []string
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(strings.ToUpper(trim), "SUMMARY:"):
			section = "summary"
			out.Summary = strings.TrimSpace(trim[len("SUMMARY:"):])
		case strings.HasPrefix(strings.ToUpper(trim), "ROOT_CAUSE:"):
			section = "root"
			out.RootCause = strings.TrimSpace(trim[len("ROOT_CAUSE:"):])
		case strings.HasPrefix(strings.ToUpper(trim), "FIX_STEPS:"):
			section = "fix"
		case strings.HasPrefix(strings.ToUpper(trim), "CONFIDENCE:"):
			section = "confidence"
			out.Confidence = strings.TrimSpace(trim[len("CONFIDENCE:"):])
		case section == "summary" && out.Summary != "" && trim != "":
			out.Summary += " " + trim
		case section == "root" && trim != "":
			if out.RootCause != "" {
				out.RootCause += " "
			}
			out.RootCause += trim
		case section == "fix" && trim != "":
			step := strings.TrimLeft(trim, "0123456789.) ")
			if step != "" {
				steps = append(steps, step)
			}
		}
	}
	out.FixSteps = steps
	return out
}
```


