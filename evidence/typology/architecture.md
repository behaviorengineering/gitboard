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
  slice_ui["ui"]
  slice_triage["triage"]
  slice_pruneagent["pruneagent"]
```


### Bounded contexts

| Slice | Objective | Route |
|-------|-----------|-------|
| `gitboard` | Core domain for managing git-based boards and orchestration. |  |
| `ui` | User interface projections and dashboards. |  |
| `triage` | Automated triage of repository changes. |  |
| `pruneagent` | Automated pruning of stale resources. |  |


### Context details

#### `gitboard`

Core domain for managing git-based boards and orchestration.

- Packages: ./internal/board, ./internal/config, ./internal/observability, ./internal/server, ./internal/syncproj, ./internal/git, ./cmd/gitboard
- Surfaces: gitboard-cli (cli)
- Programs: 

#### `ui`

User interface projections and dashboards.

- Packages: ./internal/ui/dashboard
- Surfaces: dashboard (ui)
- Programs: 

#### `triage`

Automated triage of repository changes.

- Packages: ./internal/triage, ./internal/llm
- Surfaces: 
- Programs: 

#### `pruneagent`

Automated pruning of stale resources.

- Packages: ./internal/pruneagent
- Surfaces: 
- Programs: 


### Declared coupling

No slice bindings are declared.


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
- unmapped package "internal/dashboard" in module; not claimed by any slice in owns[] or surfaces[]
- unmapped package "internal/localgit" in module; not claimed by any slice in owns[] or surfaces[]
- unmapped package "internal/remotegit" in module; not claimed by any slice in owns[] or surfaces[]
- `gitboard`: SliceBinding gitboard -> pruneagent missing but cross-slice import exists (cmd-gitboard -> internal-pruneagent)
- `gitboard`: SliceBinding gitboard -> pruneagent missing but cross-slice import exists (internal-server -> internal-pruneagent)
- `gitboard`: SliceBinding gitboard -> triage missing but cross-slice import exists (cmd-gitboard -> internal-llm)
- `gitboard`: SliceBinding gitboard -> triage missing but cross-slice import exists (cmd-gitboard -> internal-triage)
- `gitboard`: SliceBinding gitboard -> triage missing but cross-slice import exists (internal-server -> internal-triage)
- `gitboard`: source package "./internal/git" not found for component "internal-git"
- `pruneagent`: SliceBinding pruneagent -> triage missing but cross-slice import exists (internal-pruneagent -> internal-llm)
- `triage`: SliceBinding triage -> gitboard missing but cross-slice import exists (internal-llm -> internal-config)
- `triage`: SliceBinding triage -> gitboard missing but cross-slice import exists (internal-triage -> internal-config)
- `ui`: source package "./internal/ui/dashboard" not found for component "internal-dashboard"


## Agent review protocol

1. Read the relevant catalog rows and the evidence named by each finding.
2. Fix the code or catalog when the boundary is wrong.
3. Record a temporary boundary decision in the Typology journey when the design needs a later refactor.
4. Re-run `typology architecture REPO` and `typology validate REPO`.
5. Remove the generated marker only after a human accepts the narrative as a reviewed explanation.
