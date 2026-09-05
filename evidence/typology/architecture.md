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
  slice_config["config"]
```


### Bounded contexts

| Slice | Objective | Route |
|-------|-----------|-------|
| `gitboard` | Primary application entry point and orchestrator for gitboard operations. |  |
| `config` | Centralized configuration management for the gitboard ecosystem. |  |


### Context details

#### `gitboard`

Primary application entry point and orchestrator for gitboard operations.

- Packages: ./internal/board, ./internal/config, ./internal/git/local, ./internal/git/remote, ./internal/ai, ./internal/triage, ./internal/pruneagent, ./cmd/gitboard, ./internal/server, ./internal/dashboard
- Surfaces: gitboard-cli (cli), gitboard-server (api), gitboard-ui (ui)
- Programs: _(none)_

#### `config`

Centralized configuration management for the gitboard ecosystem.

- Packages: 
- Surfaces: _(none)_
- Programs: _(none)_


### Declared coupling

No slice bindings are declared.


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

- unmapped package "internal/cliexec" in module; not claimed by any slice in owns[] or surfaces[]
- unmapped package "internal/llm" in module; not claimed by any slice in owns[] or surfaces[]
- unmapped package "internal/localgit" in module; not claimed by any slice in owns[] or surfaces[]
- unmapped package "internal/observability" in module; not claimed by any slice in owns[] or surfaces[]
- unmapped package "internal/remotegit" in module; not claimed by any slice in owns[] or surfaces[]
- unmapped package "internal/syncproj" in module; not claimed by any slice in owns[] or surfaces[]
- `gitboard`: source package "./internal/ai" not found for component "internal-ai"
- `gitboard`: source package "./internal/git/local" not found for component "internal-git-local"
- `gitboard`: source package "./internal/git/remote" not found for component "internal-git-remote"


## Agent review protocol

1. Read the relevant catalog rows and the evidence named by each finding.
2. Fix the code or catalog when the boundary is wrong.
3. Record a temporary boundary decision in the Typology journey when the design needs a later refactor.
4. Re-run `typology architecture REPO` and `typology validate REPO`.
5. Remove the generated marker only after a human accepts the narrative as a reviewed explanation.
