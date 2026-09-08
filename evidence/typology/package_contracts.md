# Package public contracts

Exported symbols and delivery facts from static Go analysis.
Use packageDoc, methods, jsonTags, goEmbed, and deliveryHint to place packages under owns[] vs surfaces[] (kind: cli requires a real CLI/main delivery package).

## ./cmd/gitboard
- package: `main`
- packageDoc: Local code-change board for GitLab and GitHub (127.0.0.1 only).
- hasMain: true
- jsonTags: false
- goEmbed: false
- importsNetHTTP: true
- deliveryHint: cli
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
- deliveryHint: dto
- exportedDecls: AppearanceStandalone, AppearanceSubmodule, BranchOriginSync, BranchRef, CIStatus, Dashboard, FailedJob, LocalAppearance, LocalStatus, LocalWorktree, MergedReview, OpenItems, ProjectSummary, PruneLikely, PruneSafe, Tooling
- exportedFuncs: (none)
- exportedMethods: (none)

## ./internal/cliexec
- package: `cliexec`
- hasMain: false
- jsonTags: false
- goEmbed: false
- importsNetHTTP: false
- exportedDecls: Exec, Runner
- exportedFuncs: New
- exportedMethods: Runner.LookPath, Runner.Run, Runner.RunJSON

## ./internal/config
- package: `config`
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
- exportedDecls: DefaultExample, DefaultFetchSeconds, DefaultHeadsSeconds, DefaultLLMModel, DefaultMergedSeconds, DefaultPollSeconds, FailureDump, File, GitHubSync, GitLabSync, Host, HostGitHub, HostGitLab, LLM, Local, OpenInference, Project, SyncSources, UI, Upstream
- exportedFuncs: AgentsDir, DefaultPath, Dir, Init, Load, Save
- exportedMethods: File.EffectiveFetchSeconds, File.EffectiveHeadsSeconds, File.EffectiveLLM, File.EffectiveMergedSeconds, File.EffectiveOpenInference, File.EffectivePollSeconds, Project.OpenURL, SyncSources.HasSyncSources

## ./internal/dashboard
- package: `dashboard`
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
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
- exportedDecls: Client
- exportedFuncs: New
- exportedMethods: Client.Chat, Client.Enabled

## ./internal/localgit
- package: `localgit`
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
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
- exportedDecls: Card, Evidence, Request, Result, Service, SessionStore, VerdictAskUser, VerdictDrop, VerdictKeep
- exportedFuncs: New, NewWithStore
- exportedMethods: Service.Investigate

## ./internal/remotegit
- package: `remotegit`
- hasMain: false
- jsonTags: false
- goEmbed: false
- importsNetHTTP: false
- exportedDecls: Client, GitHub, GitLab, HeadsSnapshot, RemoteHead, RepoRef, SummaryOpts, TTLCache
- exportedFuncs: EnrichPruneHints, NewGitHub, NewGitLab, NewTTLCache
- exportedMethods: GitHub.AuthStatus, GitHub.FailedJobs, GitHub.JobLog, GitHub.ListOrgRepos, GitHub.ProjectSummary, GitLab.AuthStatus, GitLab.FailedJobs, GitLab.JobLog, GitLab.ListGroupRepos, GitLab.ProjectSummary, TTLCache.Clear, TTLCache.GetOrLoadHeads, TTLCache.GetOrLoadMerged, TTLCache.SetNow

## ./internal/server
- package: `server`
- hasMain: false
- jsonTags: false
- goEmbed: true
- importsNetHTTP: true
- deliveryHint: http-surface
- exportedDecls: Options
- exportedFuncs: NewMux
- exportedMethods: (none)

## ./internal/syncproj
- package: `syncproj`
- hasMain: false
- jsonTags: false
- goEmbed: false
- importsNetHTTP: false
- exportedDecls: Candidate, ForgeLister, Lister
- exportedFuncs: AddProject, ApplySelection, Discover, FormatCandidates, ParseSelection, ProjectFromRef, PromptSyncSources, RemoveProject, SelectInteractive
- exportedMethods: ForgeLister.ListGitHub, ForgeLister.ListGitLab

## ./internal/triage
- package: `triage`
- hasMain: false
- jsonTags: true
- goEmbed: false
- importsNetHTTP: false
- exportedDecls: Analyzer, Request, Response
- exportedFuncs: New
- exportedMethods: Analyzer.Analyze, Analyzer.Enabled

