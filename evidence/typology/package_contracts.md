# Package public contracts

Exported symbols and delivery facts from static Go analysis.
Use packageDoc, methods, jsonTags, goEmbed, deliveryHint, and observed role (never folder names) to classify packages.

## ./cmd/gitboard
- package: `main`
- packageDoc: Local code-change board for GitLab and GitHub (127.0.0.1 only).
- hasMain: true
- jsonTags: false
- goEmbed: false
- importsNetHTTP: true
- importsOsExec: false
- deliveryHint: cli
- role: entrypoint
- confidence: 0.90
- roleEvidence: has_main
- exportedDecls: (none)
- exportedFuncs: (none)
- exportedMethods: (none)

## ./internal/board
- package: `board`
- packageDoc: Package board contains the JSON data types shared across the dashboard UI, forge adapters, and server handlers.
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: false
- deliveryHint: dto
- role: dto
- confidence: 0.90
- roleEvidence: json_tags, no_exported_funcs, no_exported_methods, shared_dto_multi_importer
- exportedDecls: AppearanceStandalone, AppearanceSubmodule, BranchOriginSync, BranchRef, CIStatus, Dashboard, FailedJob, LocalAppearance, LocalStatus, LocalWorktree, MergedReview, OpenItems, ProjectSummary, PruneLikely, PruneSafe, Tooling
- exportedFuncs: (none)
- exportedMethods: (none)

## ./internal/cliexec
- package: `cliexec`
- hasMain: false
- jsonTags: false
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: true
- role: exec_runner
- confidence: 0.80
- roleEvidence: imports_os_exec
- exportedDecls: Exec, Runner
- exportedFuncs: New
- exportedMethods: Runner.LookPath, Runner.Run, Runner.RunJSON

## ./internal/config
- package: `config`
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: false
- role: config
- confidence: 0.80
- roleEvidence: imports_yaml, yaml_tags, json_tags
- exportedDecls: DefaultExample, DefaultFetchSeconds, DefaultHeadsSeconds, DefaultLLMModel, DefaultMergedSeconds, DefaultPollSeconds, FailureDump, File, GitHubSync, GitLabSync, Host, HostGitHub, HostGitLab, LLM, Local, OpenInference, Project, SyncSources, UI, Upstream
- exportedFuncs: AgentsDir, DefaultPath, Dir, Init, Load, Save
- exportedMethods: File.EffectiveFetchSeconds, File.EffectiveHeadsSeconds, File.EffectiveLLM, File.EffectiveMergedSeconds, File.EffectiveOpenInference, File.EffectivePollSeconds, Project.OpenURL, SyncSources.HasSyncSources

## ./internal/dashboard
- package: `dashboard`
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: false
- role: aggregator
- confidence: 0.80
- roleEvidence: exported_logic, domain_imports_ge_1
- exportedDecls: BadRequestError, Commands, LocalGit, PruneSafeRequest, PullFFRequest, PullFFResult, Service
- exportedFuncs: ClientFor, FindProject, IsBadRequest, New, NewCommands
- exportedMethods: BadRequestError.Error, Commands.PruneSafe, Commands.PullFF, Commands.RequireMappedPath, Service.ClearCaches, Service.Collect

## ./internal/llm
- package: `llm`
- packageDoc: Package llm provides an OpenAI-compatible chat completions client.
- hasMain: false
- jsonTags: false
- goEmbed: false
- importsNetHTTP: true
- importsOsExec: false
- role: adapter
- confidence: 0.80
- roleEvidence: imports_net_http, not_http_surface, domain_imports_eq_0
- exportedDecls: Client
- exportedFuncs: New
- exportedMethods: Client.Chat, Client.Enabled

## ./internal/localgit
- package: `localgit`
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: false
- role: adapter
- confidence: 0.80
- roleEvidence: imports_exec_runner, domain_imports_eq_0
- exportedDecls: BranchSync, Checkout, Discovery, ErrDirtyTree, ErrDiverged, ErrInvalidBranch, ErrMissingBranch, ErrUpToDate, Inspector, OriginFetchCache, RemoteRef, RoleStandalone, RoleSubmodule, Status, Worktree
- exportedFuncs: DisplayID, ExpandPath, FillCheckoutMeta, InvalidateOriginSync, NewInspector, NewOriginFetchCache, ParseRemoteURL, PickPrimary, TrackKey, ValidateBranchName
- exportedMethods: Discovery.Appearances, Inspector.CommonGitDir, Inspector.ContentOnDefault, Inspector.EnrichOriginSync, Inspector.FetchOrigin, Inspector.FetchOriginCached, Inspector.InspectPath, Inspector.OriginRemote, Inspector.PullFFOnly, Inspector.RemoveSafeCheckout, Inspector.ScanRoots, Inspector.Superproject, OriginFetchCache.Clear, OriginFetchCache.MarkSuccess, OriginFetchCache.NeedsFetch, OriginFetchCache.SetNow

## ./internal/observability
- package: `observability`
- packageDoc: Package observability bootstraps OTEL and writes error-only inference dumps (same pattern as content-pipelines).
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: false
- role: observability
- confidence: 0.90
- roleEvidence: imports_otel
- exportedDecls: InitConfig
- exportedFuncs: Init, Shutdown
- exportedMethods: failureDumpProcessor.ForceFlush, failureDumpProcessor.OnEnd, failureDumpProcessor.OnStart, failureDumpProcessor.Shutdown

## ./internal/pruneagent
- package: `pruneagent`
- packageDoc: Package pruneagent investigates local-only / likely-removable branches.
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: false
- role: aggregator
- confidence: 0.80
- roleEvidence: exported_logic, domain_imports_ge_1
- exportedDecls: Card, Evidence, Request, Result, Service, SessionStore, VerdictAskUser, VerdictDrop, VerdictKeep
- exportedFuncs: New, NewWithStore
- exportedMethods: Service.Investigate

## ./internal/remotegit
- package: `remotegit`
- hasMain: false
- jsonTags: false
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: false
- role: adapter
- confidence: 0.80
- roleEvidence: imports_exec_runner, domain_imports_eq_0
- exportedDecls: Client, GitHub, GitLab, HeadsSnapshot, RemoteHead, RepoRef, SummaryOpts, TTLCache
- exportedFuncs: EnrichPruneHints, NewGitHub, NewGitLab, NewTTLCache
- exportedMethods: GitHub.AuthStatus, GitHub.FailedJobs, GitHub.JobLog, GitHub.ListOrgRepos, GitHub.ProjectSummary, GitLab.AuthStatus, GitLab.FailedJobs, GitLab.JobLog, GitLab.ListGroupRepos, GitLab.ProjectSummary, TTLCache.Clear, TTLCache.GetOrLoadHeads, TTLCache.GetOrLoadMerged, TTLCache.SetNow

## ./internal/server
- package: `server`
- hasMain: false
- jsonTags: false
- goEmbed: true
- importsNetHTTP: true
- importsOsExec: false
- deliveryHint: http-surface
- role: http_surface
- confidence: 0.90
- roleEvidence: go_embed, embeds_static, imports_net_http
- exportedDecls: Options
- exportedFuncs: NewMux
- exportedMethods: (none)

## ./internal/syncproj
- package: `syncproj`
- hasMain: false
- jsonTags: false
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: false
- role: aggregator
- confidence: 0.80
- roleEvidence: exported_logic, domain_imports_ge_1
- exportedDecls: Candidate, ForgeLister, Lister
- exportedFuncs: AddProject, ApplySelection, Discover, FormatCandidates, ParseSelection, ProjectFromRef, PromptSyncSources, RemoveProject, SelectInteractive
- exportedMethods: ForgeLister.ListGitHub, ForgeLister.ListGitLab

## ./internal/triage
- package: `triage`
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
- importsOsExec: false
- role: aggregator
- confidence: 0.80
- roleEvidence: exported_logic, domain_imports_ge_1
- exportedDecls: Analyzer, Request, Response
- exportedFuncs: New
- exportedMethods: Analyzer.Analyze, Analyzer.Enabled

