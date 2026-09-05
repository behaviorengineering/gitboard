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
  slice_git["git"]
  slice_pruneagent["pruneagent"]
  slice_triage["triage"]
  slice_dashboard -->|reads| slice_board
  slice_dashboard -->|reads| slice_config
  slice_dashboard -->|reads| slice_git
  slice_git -->|reads| slice_cliexec
  slice_git -->|reads| slice_board
  slice_git -->|reads| slice_config
  slice_llm -->|reads| slice_config
  slice_pruneagent -->|reads| slice_cliexec
  slice_pruneagent -->|reads| slice_llm
  slice_pruneagent -->|reads| slice_git
  slice_triage -->|reads| slice_config
  slice_triage -->|reads| slice_llm
  slice_gitboard -->|reads| slice_cliexec
  slice_gitboard -->|reads| slice_config
  slice_gitboard -->|reads| slice_dashboard
  slice_gitboard -->|reads| slice_llm
  slice_gitboard -->|reads| slice_git
  slice_gitboard -->|reads| slice_pruneagent
  slice_gitboard -->|reads| slice_triage
```


### Bounded contexts

| Slice | Objective | Route |
|-------|-----------|-------|
| `board` | Provide board functionality |  |
| `cliexec` | Provide command-line execution capabilities |  |
| `config` | Provide configuration management |  |
| `dashboard` | Provide dashboard functionality |  |
| `gitboard` | Provide gitboard command-line tool and related services |  |
| `llm` | Provide LLM integration capabilities |  |
| `git` | Provide git infrastructure services |  |
| `pruneagent` | Provide prune agent functionality |  |
| `triage` | Provide triage functionality |  |


### Context details

#### `board`

Provide board functionality

- Packages: ./internal/board
- Surfaces: _(none)_
- Programs: _(none)_

#### `cliexec`

Provide command-line execution capabilities

- Packages: ./internal/cliexec
- Surfaces: cliexec-cli (cli)
- Programs: _(none)_

#### `config`

Provide configuration management

- Packages: ./internal/config
- Surfaces: _(none)_
- Programs: _(none)_

#### `dashboard`

Provide dashboard functionality

- Packages: ./internal/dashboard
- Surfaces: dashboard-ui (ui)
- Programs: _(none)_

#### `gitboard`

Provide gitboard command-line tool and related services

- Packages: ./internal/observability, ./internal/server, ./internal/syncproj, ./cmd/gitboard
- Surfaces: gitboard-cli (cli)
- Programs: _(none)_

#### `llm`

Provide LLM integration capabilities

- Packages: ./internal/llm
- Surfaces: _(none)_
- Programs: _(none)_

#### `git`

Provide git infrastructure services

- Packages: ./internal/localgit, ./internal/remotegit
- Surfaces: _(none)_
- Programs: _(none)_

#### `pruneagent`

Provide prune agent functionality

- Packages: ./internal/pruneagent
- Surfaces: _(none)_
- Programs: _(none)_

#### `triage`

Provide triage functionality

- Packages: ./internal/triage
- Surfaces: _(none)_
- Programs: _(none)_


### Declared coupling

| From | To | Kind |
|------|----|------|
| `dashboard` | `board` | reads |
| `dashboard` | `config` | reads |
| `dashboard` | `git` | reads |
| `git` | `cliexec` | reads |
| `git` | `board` | reads |
| `git` | `config` | reads |
| `llm` | `config` | reads |
| `pruneagent` | `cliexec` | reads |
| `pruneagent` | `llm` | reads |
| `pruneagent` | `git` | reads |
| `triage` | `config` | reads |
| `triage` | `llm` | reads |
| `gitboard` | `cliexec` | reads |
| `gitboard` | `config` | reads |
| `gitboard` | `dashboard` | reads |
| `gitboard` | `llm` | reads |
| `gitboard` | `git` | reads |
| `gitboard` | `pruneagent` | reads |
| `gitboard` | `triage` | reads |


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
