# Typology Architecture Brief

<!-- typology:generated -->

This document is a human-readable projection of the confirmed Typology catalog and the observed Go repository. The catalog remains the machine source of truth. This brief helps people inspect whether the design matches the code.

## How to read this brief

The **intended architecture** comes from the catalog. The **observed topology** comes from the Go import graph. Findings name evidence that needs an agent or architect to fix or record as boundary debt. Typology does not infer a final design decision from a finding.

## Intended architecture

### Bounded-context map

```mermaid
flowchart LR
  slice_gitboard["gitboard"]
  slice_board["board"]
  slice_config["config"]
  slice_dashboard["dashboard"]
  slice_llm["llm"]
  slice_localgit["localgit"]
  slice_pruneagent["pruneagent"]
  slice_remotegit["remotegit"]
  slice_syncproj["syncproj"]
  slice_triage["triage"]
  slice_gitboard -->|reads| slice_config
  slice_gitboard -->|reads| slice_dashboard
  slice_gitboard -->|reads| slice_llm
  slice_gitboard -->|reads| slice_localgit
  slice_gitboard -->|reads| slice_pruneagent
  slice_gitboard -->|reads| slice_remotegit
  slice_gitboard -->|reads| slice_triage
  slice_dashboard -->|reads| slice_board
  slice_dashboard -->|reads| slice_config
  slice_dashboard -->|reads| slice_localgit
  slice_dashboard -->|reads| slice_remotegit
  slice_remotegit -->|reads| slice_board
  slice_remotegit -->|reads| slice_config
  slice_llm -->|reads| slice_config
  slice_pruneagent -->|reads| slice_llm
  slice_pruneagent -->|reads| slice_localgit
  slice_triage -->|reads| slice_config
  slice_triage -->|reads| slice_llm
```


### Bounded contexts

| Slice | Objective | Route |
|-------|-----------|-------|
| `gitboard` | Provides the primary CLI entry point and orchestrates the gitboard application ecosystem. |  |
| `board` | Manages the core representation of project status and branch history. |  |
| `config` | Provides centralized configuration management for all system components. |  |
| `dashboard` | Serves a web-based interface for visualizing project and branch data. |  |
| `llm` | Provides LLM client capabilities for intelligent triage and analysis. |  |
| `localgit` | Interfaces with the local filesystem to inspect and manipulate git worktrees. |  |
| `pruneagent` | Automates the decision-making process for cleaning up stale branches. |  |
| `remotegit` | Manages interactions with remote git providers like GitHub and GitLab. |  |
| `syncproj` | Synchronizes local project knowledge with remote repository states. |  |
| `triage` | Analyzes repository state to provide actionable insights using LLM capabilities. |  |


### Context details

#### `gitboard`

Provides the primary CLI entry point and orchestrates the gitboard application ecosystem.

- Packages: ./internal/cliexec, ./internal/observability, ./internal/syncproj, ./cmd/gitboard, ./internal/server
- Surfaces: gitboard-cli (cli)
- Programs: _(none)_

#### `board`

Manages the core representation of project status and branch history.

- Packages: ./internal/board
- Surfaces: _(none)_
- Programs: _(none)_

#### `config`

Provides centralized configuration management for all system components.

- Packages: ./internal/config
- Surfaces: _(none)_
- Programs: _(none)_

#### `dashboard`

Serves a web-based interface for visualizing project and branch data.

- Packages: ./internal/dashboard
- Surfaces: dashboard-ui (ui)
- Programs: _(none)_

#### `llm`

Provides LLM client capabilities for intelligent triage and analysis.

- Packages: ./internal/llm
- Surfaces: _(none)_
- Programs: _(none)_

#### `localgit`

Interfaces with the local filesystem to inspect and manipulate git worktrees.

- Packages: ./internal/localgit
- Surfaces: _(none)_
- Programs: _(none)_

#### `pruneagent`

Automates the decision-making process for cleaning up stale branches.

- Packages: ./internal/pruneagent
- Surfaces: _(none)_
- Programs: _(none)_

#### `remotegit`

Manages interactions with remote git providers like GitHub and GitLab.

- Packages: ./internal/remotegit
- Surfaces: _(none)_
- Programs: _(none)_

#### `syncproj`

Synchronizes local project knowledge with remote repository states.

- Packages: 
- Surfaces: _(none)_
- Programs: _(none)_

#### `triage`

Analyzes repository state to provide actionable insights using LLM capabilities.

- Packages: ./internal/triage
- Surfaces: _(none)_
- Programs: _(none)_


### Declared coupling

| From | To | Kind |
|------|----|------|
| `gitboard` | `config` | reads |
| `gitboard` | `dashboard` | reads |
| `gitboard` | `llm` | reads |
| `gitboard` | `localgit` | reads |
| `gitboard` | `pruneagent` | reads |
| `gitboard` | `remotegit` | reads |
| `gitboard` | `triage` | reads |
| `dashboard` | `board` | reads |
| `dashboard` | `config` | reads |
| `dashboard` | `localgit` | reads |
| `dashboard` | `remotegit` | reads |
| `remotegit` | `board` | reads |
| `remotegit` | `config` | reads |
| `llm` | `config` | reads |
| `pruneagent` | `llm` | reads |
| `pruneagent` | `localgit` | reads |
| `triage` | `config` | reads |
| `triage` | `llm` | reads |


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

- `localgit`: SliceBinding localgit -> gitboard missing but cross-slice import exists (internal-localgit -> internal-cliexec)
- `pruneagent`: SliceBinding pruneagent -> gitboard missing but cross-slice import exists (internal-pruneagent -> internal-cliexec)
- `remotegit`: SliceBinding remotegit -> gitboard missing but cross-slice import exists (internal-remotegit -> internal-cliexec)


## Agent review protocol

1. Read the relevant catalog rows and the evidence named by each finding.
2. Fix the code or catalog when the boundary is wrong.
3. Record a temporary boundary decision in the Typology journey when the design needs a later refactor.
4. Re-run `typology architecture REPO` and `typology validate REPO`.
5. Remove the generated marker only after a human accepts the narrative as a reviewed explanation.
