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
  slice_board["board"]
  slice_cliexec["cliexec"]
  slice_dashboard["dashboard"]
  slice_gitboard["gitboard"]
  slice_llm["llm"]
  slice_localgit["localgit"]
  slice_remotegit["remotegit"]
  slice_pruneagent["pruneagent"]
  slice_triage["triage"]
  lib_config[("config")]
  slice_dashboard -->|reads| slice_board
  slice_dashboard -->|reads| lib_config
  slice_dashboard -->|reads| slice_localgit
  slice_dashboard -->|reads| slice_remotegit
  slice_localgit -->|reads| slice_cliexec
  slice_triage -->|reads| lib_config
  slice_triage -->|reads| slice_llm
  slice_llm -->|reads| lib_config
  slice_pruneagent -->|reads| slice_cliexec
  slice_pruneagent -->|reads| slice_llm
  slice_pruneagent -->|reads| slice_localgit
  slice_remotegit -->|reads| slice_board
  slice_remotegit -->|reads| slice_cliexec
  slice_remotegit -->|reads| lib_config
  slice_gitboard -->|reads| slice_cliexec
  slice_gitboard -->|reads| lib_config
  slice_gitboard -->|reads| slice_dashboard
  slice_gitboard -->|reads| slice_llm
  slice_gitboard -->|reads| slice_localgit
  slice_gitboard -->|reads| slice_pruneagent
  slice_gitboard -->|reads| slice_remotegit
  slice_gitboard -->|reads| slice_triage
```


### Bounded contexts

| Slice | Objective | Route |
|-------|-----------|-------|
| `board` | Provides shared data structures for dashboard state and forge synchronization. |  |
| `cliexec` | Provides a standardized interface for executing system commands and shell processes. |  |
| `dashboard` | Orchestrates data from adapters and local state to provide a unified project view. |  |
| `gitboard` | Serves as the primary entrypoint for the Gitboard CLI and local dashboard server. |  |
| `llm` | Provides an OpenAI-compatible client for automated code and triage analysis. |  |
| `localgit` | Acts as an adapter for performing local git filesystem operations and inspections. |  |
| `remotegit` | Acts as an adapter for interacting with GitHub and GitLab remote APIs. |  |
| `pruneagent` | Executes automated investigations into removable local branches. |  |
| `triage` | Provides DTOs and logic for analyzing and summarizing triage requests. |  |


### Libraries

Technical packages with no domain knowledge. They claim packages without a product objective.

| Library | Purpose | Packages |
|---------|---------|----------|
| `config` | Shared technical package for application configuration management. | internal/config |


### Context details

#### `board`

Provides shared data structures for dashboard state and forge synchronization.

- Packages: internal/board
- Surfaces: _(none)_
- Programs: _(none)_

#### `cliexec`

Provides a standardized interface for executing system commands and shell processes.

- Packages: internal/cliexec
- Surfaces: _(none)_
- Programs: _(none)_

#### `dashboard`

Orchestrates data from adapters and local state to provide a unified project view.

- Packages: internal/dashboard
- Surfaces: _(none)_
- Programs: _(none)_

#### `gitboard`

Serves as the primary entrypoint for the Gitboard CLI and local dashboard server.

- Packages: internal/observability, internal/syncproj, cmd/gitboard, internal/server
- Surfaces: gitboard-cli (cli), gitboard-server (api)
- Programs: _(none)_

#### `llm`

Provides an OpenAI-compatible client for automated code and triage analysis.

- Packages: internal/llm
- Surfaces: _(none)_
- Programs: _(none)_

#### `localgit`

Acts as an adapter for performing local git filesystem operations and inspections.

- Packages: internal/localgit
- Surfaces: _(none)_
- Programs: _(none)_

#### `remotegit`

Acts as an adapter for interacting with GitHub and GitLab remote APIs.

- Packages: internal/remotegit
- Surfaces: _(none)_
- Programs: _(none)_

#### `pruneagent`

Executes automated investigations into removable local branches.

- Packages: internal/pruneagent
- Surfaces: _(none)_
- Programs: _(none)_

#### `triage`

Provides DTOs and logic for analyzing and summarizing triage requests.

- Packages: internal/triage
- Surfaces: _(none)_
- Programs: _(none)_


### Declared coupling

| From | To | Kind |
|------|----|------|
| `dashboard` | `board` | reads |
| `dashboard` | `config` | reads |
| `dashboard` | `localgit` | reads |
| `dashboard` | `remotegit` | reads |
| `localgit` | `cliexec` | reads |
| `triage` | `config` | reads |
| `triage` | `llm` | reads |
| `llm` | `config` | reads |
| `pruneagent` | `cliexec` | reads |
| `pruneagent` | `llm` | reads |
| `pruneagent` | `localgit` | reads |
| `remotegit` | `board` | reads |
| `remotegit` | `cliexec` | reads |
| `remotegit` | `config` | reads |
| `gitboard` | `cliexec` | reads |
| `gitboard` | `config` | reads |
| `gitboard` | `dashboard` | reads |
| `gitboard` | `llm` | reads |
| `gitboard` | `localgit` | reads |
| `gitboard` | `pruneagent` | reads |
| `gitboard` | `remotegit` | reads |
| `gitboard` | `triage` | reads |


Bindings may target a slice or a library. `from` is always a slice.

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

No catalog, path, import, or documentation findings were reported.


## Agent review protocol

1. Read the relevant catalog rows and the evidence named by each finding.
2. Fix the code or catalog when the boundary is wrong.
3. Record a temporary boundary decision in the Typology journey when the design needs a later refactor.
4. Re-run `typology architecture REPO` and `typology validate REPO`.
5. Remove the generated marker only after a human accepts the narrative as a reviewed explanation.
