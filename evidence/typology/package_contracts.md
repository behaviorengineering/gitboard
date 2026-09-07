# Package public contracts

Exported symbols and main entrypoints from static Go analysis.
Use this to place packages under owns[] vs surfaces[] (kind: cli requires a real CLI/main delivery package).

## ./cmd/gitboard
- package: `main`
- hasMain: true
- exportedDecls: (none)
- exportedFuncs: (none)

## ./internal/board
- package: `board`
- hasMain: false
- exportedDecls: AppearanceStandalone, AppearanceSubmodule, BranchOriginSync, BranchRef, CIStatus, Dashboard, FailedJob, LocalAppearance, LocalStatus, LocalWorktree, MergedReview, OpenItems, ProjectSummary, PruneLikely, PruneSafe, Tooling
- exportedFuncs: (none)

## ./internal/cliexec
- package: `cliexec`
- hasMain: false
- exportedDecls: Exec, Runner
- exportedFuncs: New

## ./internal/config
- package: `config`
- hasMain: false
- exportedDecls: DefaultExample, DefaultFetchSeconds, DefaultHeadsSeconds, DefaultLLMModel, DefaultMergedSeconds, DefaultPollSeconds, FailureDump, File, GitHubSync, GitLabSync, Host, HostGitHub, HostGitLab, LLM, Local, OpenInference, Project, SyncSources, UI, Upstream
- exportedFuncs: AgentsDir, DefaultPath, Dir, Init, Load, Save

## ./internal/dashboard
- package: `dashboard`
- hasMain: false
- exportedDecls: BadRequestError, Commands, LocalGit, PruneSafeRequest, PullFFRequest, PullFFResult, Service
- exportedFuncs: ClientFor, FindProject, IsBadRequest, New, NewCommands

## ./internal/llm
- package: `llm`
- hasMain: false
- exportedDecls: Client
- exportedFuncs: New

## ./internal/localgit
- package: `localgit`
- hasMain: false
- exportedDecls: BranchSync, Checkout, Discovery, ErrDirtyTree, ErrDiverged, ErrInvalidBranch, ErrMissingBranch, ErrUpToDate, Inspector, OriginFetchCache, RemoteRef, RoleStandalone, RoleSubmodule, Status, Worktree
- exportedFuncs: DisplayID, ExpandPath, FillCheckoutMeta, InvalidateOriginSync, NewInspector, NewOriginFetchCache, ParseRemoteURL, PickPrimary, TrackKey, ValidateBranchName

## ./internal/observability
- package: `observability`
- hasMain: false
- exportedDecls: InitConfig
- exportedFuncs: Init, Shutdown

## ./internal/pruneagent
- package: `pruneagent`
- hasMain: false
- exportedDecls: Card, Evidence, Request, Result, Service, SessionStore, VerdictAskUser, VerdictDrop, VerdictKeep
- exportedFuncs: New, NewWithStore

## ./internal/remotegit
- package: `remotegit`
- hasMain: false
- exportedDecls: Client, GitHub, GitLab, HeadsSnapshot, RemoteHead, RepoRef, SummaryOpts, TTLCache
- exportedFuncs: EnrichPruneHints, NewGitHub, NewGitLab, NewTTLCache

## ./internal/server
- package: `server`
- hasMain: false
- exportedDecls: Options
- exportedFuncs: NewMux

## ./internal/syncproj
- package: `syncproj`
- hasMain: false
- exportedDecls: Candidate, ForgeLister, Lister
- exportedFuncs: AddProject, ApplySelection, Discover, FormatCandidates, ParseSelection, ProjectFromRef, PromptSyncSources, RemoveProject, SelectInteractive

## ./internal/triage
- package: `triage`
- hasMain: false
- exportedDecls: Analyzer, Request, Response
- exportedFuncs: New

