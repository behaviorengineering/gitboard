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
  slice_cliexec["cliexec"]
  slice_config["config"]
  slice_dashboard["dashboard"]
  slice_gitboard["gitboard"]
  slice_llm["llm"]
  slice_localgit["localgit"]
  slice_observability["observability"]
  slice_pruneagent["pruneagent"]
  slice_remotegit["remotegit"]
  slice_server["server"]
  slice_syncproj["syncproj"]
  slice_triage["triage"]
  slice_llm -->|reads| slice_config
  slice_remotegit -->|reads| slice_board
  slice_remotegit -->|reads| slice_cliexec
  slice_remotegit -->|reads| slice_config
  slice_triage -->|reads| slice_config
  slice_triage -->|reads| slice_llm
  slice_dashboard -->|reads| slice_board
  slice_dashboard -->|reads| slice_config
  slice_dashboard -->|reads| slice_localgit
  slice_dashboard -->|reads| slice_remotegit
  slice_localgit -->|reads| slice_cliexec
  slice_pruneagent -->|reads| slice_cliexec
  slice_pruneagent -->|reads| slice_llm
  slice_pruneagent -->|reads| slice_localgit
  slice_server -->|reads| slice_config
  slice_server -->|reads| slice_dashboard
  slice_server -->|reads| slice_pruneagent
  slice_server -->|reads| slice_triage
  slice_syncproj -->|reads| slice_config
  slice_syncproj -->|reads| slice_remotegit
  slice_gitboard -->|reads| slice_cliexec
  slice_gitboard -->|reads| slice_config
  slice_gitboard -->|reads| slice_dashboard
  slice_gitboard -->|reads| slice_llm
  slice_gitboard -->|reads| slice_localgit
  slice_gitboard -->|reads| slice_observability
  slice_gitboard -->|reads| slice_pruneagent
  slice_gitboard -->|reads| slice_remotegit
  slice_gitboard -->|reads| slice_server
  slice_gitboard -->|reads| slice_syncproj
  slice_gitboard -->|reads| slice_triage
```


### Bounded contexts

| Slice | Objective | Route |
|-------|-----------|-------|
| `board` |  |  |
| `cliexec` |  |  |
| `config` |  |  |
| `dashboard` |  |  |
| `gitboard` |  |  |
| `llm` |  |  |
| `localgit` |  |  |
| `observability` |  |  |
| `pruneagent` |  |  |
| `remotegit` |  |  |
| `server` |  |  |
| `syncproj` |  |  |
| `triage` |  |  |


### Context details

#### `board`



- Packages: ./internal/board
- Surfaces: 
- Programs: 

#### `cliexec`



- Packages: ./internal/cliexec
- Surfaces: cliexec-cli (cli)
- Programs: 

#### `config`



- Packages: ./internal/config
- Surfaces: 
- Programs: 

#### `dashboard`



- Packages: ./internal/dashboard
- Surfaces: 
- Programs: 

#### `gitboard`



- Packages: ./cmd/gitboard
- Surfaces: gitboard-cli (cli)
- Programs: 

#### `llm`



- Packages: ./internal/llm
- Surfaces: 
- Programs: 

#### `localgit`



- Packages: ./internal/localgit
- Surfaces: 
- Programs: 

#### `observability`



- Packages: ./internal/observability
- Surfaces: 
- Programs: 

#### `pruneagent`



- Packages: ./internal/pruneagent
- Surfaces: 
- Programs: 

#### `remotegit`



- Packages: ./internal/remotegit
- Surfaces: 
- Programs: 

#### `server`



- Packages: ./internal/server
- Surfaces: 
- Programs: 

#### `syncproj`



- Packages: ./internal/syncproj
- Surfaces: 
- Programs: 

#### `triage`



- Packages: ./internal/triage
- Surfaces: 
- Programs: 


### Declared coupling

| From | To | Kind |
|------|----|------|
| `llm` | `config` | reads |
| `remotegit` | `board` | reads |
| `remotegit` | `cliexec` | reads |
| `remotegit` | `config` | reads |
| `triage` | `config` | reads |
| `triage` | `llm` | reads |
| `dashboard` | `board` | reads |
| `dashboard` | `config` | reads |
| `dashboard` | `localgit` | reads |
| `dashboard` | `remotegit` | reads |
| `localgit` | `cliexec` | reads |
| `pruneagent` | `cliexec` | reads |
| `pruneagent` | `llm` | reads |
| `pruneagent` | `localgit` | reads |
| `server` | `config` | reads |
| `server` | `dashboard` | reads |
| `server` | `pruneagent` | reads |
| `server` | `triage` | reads |
| `syncproj` | `config` | reads |
| `syncproj` | `remotegit` | reads |
| `gitboard` | `cliexec` | reads |
| `gitboard` | `config` | reads |
| `gitboard` | `dashboard` | reads |
| `gitboard` | `llm` | reads |
| `gitboard` | `localgit` | reads |
| `gitboard` | `observability` | reads |
| `gitboard` | `pruneagent` | reads |
| `gitboard` | `remotegit` | reads |
| `gitboard` | `server` | reads |
| `gitboard` | `syncproj` | reads |
| `gitboard` | `triage` | reads |


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

- `board`: DocPage components path missing: docs/develop/board/components.md
- `board`: DocPage overview path missing: docs/develop/board/overview.md
- `board`: missing objective
- `cliexec`: DocPage cli path missing: docs/develop/cliexec/cli.md
- `cliexec`: DocPage components path missing: docs/develop/cliexec/components.md
- `cliexec`: DocPage overview path missing: docs/develop/cliexec/overview.md
- `cliexec`: missing objective
- `config`: DocPage components path missing: docs/develop/config/components.md
- `config`: DocPage overview path missing: docs/develop/config/overview.md
- `config`: missing objective
- `dashboard`: DocPage components path missing: docs/develop/dashboard/components.md
- `dashboard`: DocPage overview path missing: docs/develop/dashboard/overview.md
- `dashboard`: missing objective
- `gitboard`: DocPage cli path missing: docs/develop/gitboard/cli.md
- `gitboard`: DocPage components path missing: docs/develop/gitboard/components.md
- `gitboard`: DocPage overview path missing: docs/develop/gitboard/overview.md
- `gitboard`: missing objective
- `llm`: DocPage components path missing: docs/develop/llm/components.md
- `llm`: DocPage overview path missing: docs/develop/llm/overview.md
- `llm`: missing objective
- `localgit`: DocPage components path missing: docs/develop/localgit/components.md
- `localgit`: DocPage overview path missing: docs/develop/localgit/overview.md
- `localgit`: missing objective
- `observability`: DocPage components path missing: docs/develop/observability/components.md
- `observability`: DocPage overview path missing: docs/develop/observability/overview.md
- `observability`: missing objective
- `pruneagent`: DocPage components path missing: docs/develop/pruneagent/components.md
- `pruneagent`: DocPage overview path missing: docs/develop/pruneagent/overview.md
- `pruneagent`: missing objective
- `remotegit`: DocPage components path missing: docs/develop/remotegit/components.md
- `remotegit`: DocPage overview path missing: docs/develop/remotegit/overview.md
- `remotegit`: missing objective
- `server`: DocPage components path missing: docs/develop/server/components.md
- `server`: DocPage overview path missing: docs/develop/server/overview.md
- `server`: missing objective
- `syncproj`: DocPage components path missing: docs/develop/syncproj/components.md
- `syncproj`: DocPage overview path missing: docs/develop/syncproj/overview.md
- `syncproj`: missing objective
- `triage`: DocPage components path missing: docs/develop/triage/components.md
- `triage`: DocPage overview path missing: docs/develop/triage/overview.md
- `triage`: missing objective


## Agent review protocol

1. Read the relevant catalog rows and the evidence named by each finding.
2. Fix the code or catalog when the boundary is wrong.
3. Record a temporary boundary decision in the Typology journey when the design needs a later refactor.
4. Re-run `typology architecture REPO` and `typology validate REPO`.
5. Remove the generated marker only after a human accepts the narrative as a reviewed explanation.
