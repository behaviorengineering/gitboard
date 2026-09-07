# Typology Architecture Brief

> **Typology seed evidence.** Post-survey/refine architecture brief for this context digest proposal. Not the teaching-story root `architecture.md`, and not the confirmed `.typology/` catalog.

<!-- typology:generated -->

This document is a human-readable projection of the confirmed Typology catalog and the observed Go repository. The catalog remains the machine source of truth. This brief helps people inspect whether the design matches the code.

## How to read this brief

The **intended architecture** comes from the catalog. The **observed topology** comes from the Go import graph. Findings name evidence that needs an agent or architect to fix or record as boundary debt. Typology does not infer a final design decision from a finding.

## Intended architecture

### Bounded-context map

```mermaid
flowchart LR
  slice_gitboard["gitboard"]
  slice_config["config"]
  slice_dashboard["dashboard"]
  slice_localgit["localgit"]
  slice_remotegit["remotegit"]
  slice_triage["triage"]
  slice_triage -->|reads| slice_config
  slice_localgit -->|reads| slice_remotegit
  slice_gitboard -->|reads| slice_config
  slice_gitboard -->|reads| slice_dashboard
  slice_gitboard -->|reads| slice_triage
  slice_dashboard -->|reads| slice_localgit
```


### Bounded contexts

| Slice | Objective | Route |
|-------|-----------|-------|
| `gitboard` | Orchestrates the complete gitboard lifecycle from local discovery to remote synchronization and dashboarding. |  |
| `config` | Provides a centralized configuration management system for all gitboard components and host settings. |  |
| `dashboard` | Serves as the visual management interface for monitoring project status and performing remote actions. |  |
| `localgit` | Manages local filesystem interactions and git worktree inspections. |  |
| `remotegit` | Facilitates interaction with remote git providers like GitHub and GitLab to synchronize state. |  |
| `triage` | Automates the analysis and pruning of git repositories using agentic workflows and LLM capabilities. |  |


### Context details

#### `gitboard`

Orchestrates the complete gitboard lifecycle from local discovery to remote synchronization and dashboarding.

- Packages: ./internal/board, ./cmd/gitboard, ./internal/observability, ./internal/server, ./internal/syncproj
- Surfaces: gitboard-cli (cli)
- Programs: _(none)_

#### `config`

Provides a centralized configuration management system for all gitboard components and host settings.

- Packages: ./internal/config
- Surfaces: _(none)_
- Programs: _(none)_

#### `dashboard`

Serves as the visual management interface for monitoring project status and performing remote actions.

- Packages: ./internal/dashboard
- Surfaces: dashboard-ui (ui)
- Programs: _(none)_

#### `localgit`

Manages local filesystem interactions and git worktree inspections.

- Packages: ./internal/localgit
- Surfaces: _(none)_
- Programs: _(none)_

#### `remotegit`

Facilitates interaction with remote git providers like GitHub and GitLab to synchronize state.

- Packages: ./internal/remotegit, ./internal/cliexec
- Surfaces: _(none)_
- Programs: _(none)_

#### `triage`

Automates the analysis and pruning of git repositories using agentic workflows and LLM capabilities.

- Packages: ./internal/triage, ./internal/pruneagent, ./internal/llm
- Surfaces: _(none)_
- Programs: _(none)_


### Declared coupling

| From | To | Kind |
|------|----|------|
| `triage` | `config` | reads |
| `localgit` | `remotegit` | reads |
| `gitboard` | `config` | reads |
| `gitboard` | `dashboard` | reads |
| `gitboard` | `triage` | reads |
| `dashboard` | `localgit` | reads |


No component bindings are declared.


## Observed topology

The repository contains **13 Go packages** in the inspected modules:
- `./cmd/gitboard`
- `./internal/board`
- `./internal/cliexec`
- `./internal/config`
- `./internal/dashboard`
- `./internal/llm`
- `./internal/localgit`
- `./internal/observability`
- `./internal/pruneagent`
- `./internal/remotegit`
- `./internal/server`
- `./internal/syncproj`
- `./internal/triage`


### High-coupling packages

| Package | In-degree | Out-degree |
|---------|-----------|------------|
| `./cmd/gitboard` | 0 | 11 |
| `./internal/cliexec` | 4 | 0 |
| `./internal/config` | 7 | 0 |
| `./internal/dashboard` | 2 | 4 |
| `./internal/llm` | 3 | 1 |
| `./internal/localgit` | 3 | 1 |
| `./internal/remotegit` | 3 | 3 |
| `./internal/server` | 1 | 4 |

### Leaf packages

Leaf packages have no outgoing local imports. They may be valid infrastructure leaves or packages that should sit inside their only caller.

| Package | Imported by |
|---------|-------------|
| `./internal/board` | 2 |
| `./internal/cliexec` | 4 |
| `./internal/config` | 7 |
| `./internal/observability` | 1 |

### Shared platform leaves

./internal/config


### Merge candidates

These are heuristics for review, not automatic moves.

- `./internal/observability` may sit with `./cmd/gitboard`: sole importer: only imported by ./cmd/gitboard
- `./internal/server` may sit with `./cmd/gitboard`: sole importer: only imported by ./cmd/gitboard
- `./internal/syncproj` may sit with `./cmd/gitboard`: sole importer: only imported by ./cmd/gitboard



## Drift and design questions

The following findings need a correction or an explicit boundary-debt decision:

- `dashboard`: SliceBinding dashboard -> config missing but cross-slice import exists (internal-dashboard -> internal-config)
- `dashboard`: SliceBinding dashboard -> gitboard missing but cross-slice import exists (internal-dashboard -> internal-board)
- `dashboard`: SliceBinding dashboard -> remotegit missing but cross-slice import exists (internal-dashboard -> internal-remotegit)
- `gitboard`: SliceBinding gitboard -> localgit missing but cross-slice import exists (cmd-gitboard -> internal-localgit)
- `gitboard`: SliceBinding gitboard -> remotegit missing but cross-slice import exists (cmd-gitboard -> internal-cliexec)
- `gitboard`: SliceBinding gitboard -> remotegit missing but cross-slice import exists (cmd-gitboard -> internal-remotegit)
- `gitboard`: SliceBinding gitboard -> remotegit missing but cross-slice import exists (internal-syncproj -> internal-remotegit)
- `remotegit`: SliceBinding remotegit -> config missing but cross-slice import exists (internal-remotegit -> internal-config)
- `remotegit`: SliceBinding remotegit -> gitboard missing but cross-slice import exists (internal-remotegit -> internal-board)
- `triage`: SliceBinding triage -> localgit missing but cross-slice import exists (internal-pruneagent -> internal-localgit)
- `triage`: SliceBinding triage -> remotegit missing but cross-slice import exists (internal-pruneagent -> internal-cliexec)


## Agent review protocol

1. Read the relevant catalog rows and the evidence named by each finding.
2. Fix the code or catalog when the boundary is wrong.
3. Record a temporary boundary decision in the Typology journey when the design needs a later refactor.
4. Re-run `typology architecture REPO` and `typology validate REPO`.
5. Remove the generated marker only after a human accepts the narrative as a reviewed explanation.
