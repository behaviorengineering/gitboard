# Typology Architecture Brief

<!-- typology:generated -->

This document is a human-readable projection of the confirmed Typology catalog and the observed Go repository. The catalog remains the machine source of truth. This brief helps people inspect whether the design matches the code.

## How to read this brief

The **intended architecture** comes from the catalog. The **observed topology** comes from the Go import graph. Findings name evidence that needs an agent or architect to fix or record as boundary debt. Typology does not infer a final design decision from a finding.

## Intended architecture

### Bounded-context map

```mermaid
flowchart LR
  slice_board["board"]
  slice_config["config"]
  slice_dashboard["dashboard"]
  slice_gitboard["gitboard"]
  slice_git_engine["git-engine"]
  slice_gitboard_llm_gateway["gitboard-llm-gateway"]
  slice_gitboard -->|reads| slice_config
  slice_gitboard -->|reads| slice_dashboard
  slice_gitboard -->|reads| slice_gitboard_llm_gateway
  slice_gitboard -->|reads| slice_git_engine
  slice_gitboard -->|reads| slice_git_engine
  slice_gitboard_llm_gateway -->|reads| slice_config
  slice_dashboard -->|reads| slice_board
  slice_dashboard -->|reads| slice_config
  slice_dashboard -->|reads| slice_git_engine
  slice_git_engine -->|reads| slice_board
  slice_git_engine -->|reads| slice_config
```


### Bounded contexts

| Slice | Objective | Route |
|-------|-----------|-------|
| `board` | Core domain logic for board state management. |  |
| `config` | Centralized configuration management for the repository. |  |
| `dashboard` | Visual representation and monitoring of gitboard state. |  |
| `gitboard` | Main application entry point and orchestrator. |  |
| `git-engine` | Unified handling of local and remote git operations. |  |
| `gitboard-llm-gateway` | Capability provider for LLM-based operations. |  |


### Context details

#### `board`

Core domain logic for board state management.

- Packages: ./internal/board
- Surfaces: 
- Programs: 

#### `config`

Centralized configuration management for the repository.

- Packages: ./internal/config
- Surfaces: 
- Programs: 

#### `dashboard`

Visual representation and monitoring of gitboard state.

- Packages: ./internal/dashboard
- Surfaces: 
- Programs: 

#### `gitboard`

Main application entry point and orchestrator.

- Packages: ./internal/server, ./internal/syncproj, ./internal/observability, ./internal/pruneagent, ./internal/triage, ./cmd/gitboard
- Surfaces: gitboard-cli (cli)
- Programs: 

#### `git-engine`

Unified handling of local and remote git operations.

- Packages: ./internal/localgit, ./internal/remotegit
- Surfaces: 
- Programs: 

#### `gitboard-llm-gateway`

Capability provider for LLM-based operations.

- Packages: ./internal/llm
- Surfaces: 
- Programs: 


### Declared coupling

| From | To | Kind |
|------|----|------|
| `gitboard` | `config` | reads |
| `gitboard` | `dashboard` | reads |
| `gitboard` | `gitboard-llm-gateway` | reads |
| `gitboard` | `git-engine` | reads |
| `gitboard` | `git-engine` | reads |
| `gitboard-llm-gateway` | `config` | reads |
| `dashboard` | `board` | reads |
| `dashboard` | `config` | reads |
| `dashboard` | `git-engine` | reads |
| `git-engine` | `board` | reads |
| `git-engine` | `config` | reads |


No component bindings are declared.


## Observed topology

The repository contains **13 Go packages** in the inspected modules:
- `.`


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

- unmapped package "internal/cliexec" in module; not claimed by any slice in owns[] or surfaces[]
- `board`: DocPage components path missing: docs/develop/board/components.md
- `board`: DocPage overview path missing: docs/develop/board/overview.md
- `config`: DocPage components path missing: docs/develop/config/components.md
- `config`: DocPage overview path missing: docs/develop/config/overview.md
- `dashboard`: DocPage components path missing: docs/develop/dashboard/components.md
- `dashboard`: DocPage overview path missing: docs/develop/dashboard/overview.md
- `git-engine`: DocPage components path missing: docs/develop/localgit/components.md
- `git-engine`: DocPage overview path missing: docs/develop/localgit/overview.md
- `gitboard`: DocPage cli path missing: docs/develop/gitboard/cli.md
- `gitboard`: DocPage components path missing: docs/develop/gitboard/components.md
- `gitboard`: DocPage overview path missing: docs/develop/gitboard/overview.md
- `gitboard-llm-gateway`: DocPage components path missing: docs/develop/llm/components.md
- `gitboard-llm-gateway`: DocPage overview path missing: docs/develop/llm/overview.md


## Agent review protocol

1. Read the relevant catalog rows and the evidence named by each finding.
2. Fix the code or catalog when the boundary is wrong.
3. Record a temporary boundary decision in the Typology journey when the design needs a later refactor.
4. Re-run `typology architecture REPO` and `typology validate REPO`.
5. Remove the generated marker only after a human accepts the narrative as a reviewed explanation.
