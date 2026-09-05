# Architecture

  Gitboard is organized into bounded contexts that separate core orchestration, user interface, and automated intelligence.

  ## Bounded Contexts

  ### `gitboard` (Core Domain)
  Manages the core orchestration of git-based boards, configuration, and synchronization.
  - **Key Components**: `internal/board`, `internal/config`, `internal/server`, `internal/syncproj`, `internal/git`.
  - **Surfaces**: `gitboard-cli` (via `cmd/gitboard`).

  ### `ui` (Projections)
  Handles user interface projections and the web-based dashboard.
  - **Key Components**: `internal/ui/dashboard`.
  - **Surfaces**: Web Dashboard (port `:1325`).

  ### `triage` (Automated Intelligence)
  Provides automated triage of repository changes using LLM capabilities.
  - **Key Components**: `internal/triage`, `internal/llm`.

  ### `pruneagent` (Resource Management)
  Handles the automated pruning of stale resources.
  - **Key Components**: `internal/pruneagent`.

  ## Boundary Debt &amp; Design Notes
  - **Capability Leak**: The `llm` package currently exists as a standalone package but is intended to be consumed as a capability by `triage` and `pruneagent`.
  - **Git Domain Fragmentation**: The git domain is undergoing transition from `localgit`/`remotegit` to a unified `internal/git`.
  - **Surface/Domain Confusion**: `cliexec` is currently a package but is slated to be treated strictly as a CLI surface component of the `gitboard` slice.