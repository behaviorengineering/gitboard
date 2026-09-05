# Architecture

Gitboard is organized into three primary bounded contexts:

## Bounded Contexts

### gitboard (Core Domain)
Manages orchestration, primary interfaces, and core operations.
- **Components**: Board, Config, Observability, Server, Sync Project, Prune Agent.
- **Surfaces**: 
  - `gitboard-cli` (via `cmd/gitboard`)
  - `gitboard-ui` (via `internal/dashboard`)

### git-engine (Git Primitives)
Provides low-level primitives for local and remote repository manipulation.
- **Components**: `localgit`, `remotegit`.

### triage (Automated Workflow)
Handles the automated analysis and categorization of repository changes.
- **Components**: `triage`, `llm`.

## Boundary Map
- `git-engine` reads from `gitboard`.
- `gitboard` reads from `git-engine`.
- `triage` reads from `gitboard`.