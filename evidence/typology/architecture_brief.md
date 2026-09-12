# Typology Architecture Brief

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: README.md](README.md) · [Next: cluster_proposal.md](cluster_proposal.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

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
  slice_dto_board["dto/board"]
  slice_runner["runner"]
  slice_dashboard["dashboard"]
  slice_git["git"]
  slice_llm["llm"]
  slice_observability["observability"]
  slice_analysis["analysis"]
  slice_server["server"]
  slice_syncproj["syncproj"]
  lib_config[("config")]
  slice_gitboard -->|reads| slice_runner
  slice_gitboard -->|reads| lib_config
  slice_gitboard -->|reads| slice_dashboard
  slice_gitboard -->|reads| slice_llm
  slice_gitboard -->|reads| slice_git
  slice_gitboard -->|reads| slice_observability
  slice_gitboard -->|reads| slice_analysis
  slice_gitboard -->|reads| slice_server
  slice_gitboard -->|reads| slice_syncproj
  slice_dashboard -->|reads| slice_dto_board
  slice_dashboard -->|reads| lib_config
  slice_dashboard -->|reads| slice_git
  slice_llm -->|reads| lib_config
  slice_git -->|reads| slice_dto_board
  slice_git -->|reads| slice_runner
  slice_git -->|reads| lib_config
  slice_server -->|reads| lib_config
  slice_server -->|reads| slice_dashboard
  slice_server -->|reads| slice_analysis
  slice_syncproj -->|reads| lib_config
  slice_syncproj -->|reads| slice_git
  slice_analysis -->|reads| lib_config
  slice_analysis -->|reads| slice_llm
```


### Bounded contexts

| Slice | Objective | Route |
|-------|-----------|-------|
| `gitboard` | Acts as the command-line entrypoint to orchestrate the gitboard application. |  |
| `dto/board` | Provide JSON data types shared across the dashboard UI, forge adapters, and server handlers. |  |
| `runner` | Executes external processes via the os/exec package. |  |
| `dashboard` | The dashboard package aggregates various views and investigation results for the system. |  |
| `git` | Provides clients for interacting with remote Git forge APIs like GitHub and GitLab. |  |
| `llm` | Provides an OpenAI-compatible chat completions client. |  |
| `observability` | Bootstraps OTEL and manages error-only inference dumps. |  |
| `analysis` | Analyzes project and job data to produce triage responses. |  |
| `server` | Provides an HTTP server interface and mux routing for the application. |  |
| `syncproj` | Aggregates views for project discovery and selection. |  |


### Libraries

Technical packages with no domain knowledge. They claim packages without a product objective.

| Library | Purpose | Packages |
|---------|---------|----------|
| `config` | Shared technical package without domain knowledge | internal/config |


### Context details

#### `gitboard`

Acts as the command-line entrypoint to orchestrate the gitboard application.

- Packages: cmd/gitboard
- Surfaces: gitboard-cli (cli)
- Programs: _(none)_

#### `dto/board`

Provide JSON data types shared across the dashboard UI, forge adapters, and server handlers.

- Packages: internal/board
- Surfaces: _(none)_
- Programs: _(none)_

#### `runner`

Executes external processes via the os/exec package.

- Packages: internal/cliexec
- Surfaces: _(none)_
- Programs: _(none)_

#### `dashboard`

The dashboard package aggregates various views and investigation results for the system.

- Packages: internal/dashboard
- Surfaces: _(none)_
- Programs: _(none)_

#### `git`

Provides clients for interacting with remote Git forge APIs like GitHub and GitLab.

- Packages: internal/localgit, internal/remotegit
- Surfaces: _(none)_
- Programs: _(none)_

#### `llm`

Provides an OpenAI-compatible chat completions client.

- Packages: internal/llm
- Surfaces: _(none)_
- Programs: _(none)_

#### `observability`

Bootstraps OTEL and manages error-only inference dumps.

- Packages: internal/observability
- Surfaces: _(none)_
- Programs: _(none)_

#### `analysis`

Analyzes project and job data to produce triage responses.

- Packages: internal/pruneagent, internal/triage
- Surfaces: _(none)_
- Programs: _(none)_

#### `server`

Provides an HTTP server interface and mux routing for the application.

- Packages: internal/server
- Surfaces: server-ui (ui)
- Programs: _(none)_

#### `syncproj`

Aggregates views for project discovery and selection.

- Packages: internal/syncproj
- Surfaces: _(none)_
- Programs: _(none)_


### Declared coupling

| From | To | Kind |
|------|----|------|
| `gitboard` | `runner` | reads |
| `gitboard` | `config` | reads |
| `gitboard` | `dashboard` | reads |
| `gitboard` | `llm` | reads |
| `gitboard` | `git` | reads |
| `gitboard` | `observability` | reads |
| `gitboard` | `analysis` | reads |
| `gitboard` | `server` | reads |
| `gitboard` | `syncproj` | reads |
| `dashboard` | `dto/board` | reads |
| `dashboard` | `config` | reads |
| `dashboard` | `git` | reads |
| `llm` | `config` | reads |
| `git` | `dto/board` | reads |
| `git` | `runner` | reads |
| `git` | `config` | reads |
| `server` | `config` | reads |
| `server` | `dashboard` | reads |
| `server` | `analysis` | reads |
| `syncproj` | `config` | reads |
| `syncproj` | `git` | reads |
| `analysis` | `config` | reads |
| `analysis` | `llm` | reads |


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

The following findings need a correction or an explicit boundary-debt decision:

- `analysis`: SliceBinding analysis -> git missing but cross-slice import exists (internal-pruneagent -> internal-localgit)
- `analysis`: SliceBinding analysis -> runner missing but cross-slice import exists (internal-pruneagent -> internal-cliexec)


## Agent review protocol

1. Read the relevant catalog rows and the evidence named by each finding.
2. Fix the code or catalog when the boundary is wrong.
3. Record a temporary boundary decision in the Typology journey when the design needs a later refactor.
4. Re-run `typology architecture REPO` and `typology validate REPO`.
5. Remove the generated marker only after a human accepts the narrative as a reviewed explanation.
