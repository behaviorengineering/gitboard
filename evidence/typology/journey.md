# Journey: Gitboard Typology Refinement

**Status:** Open (Awaiting Human Decision on Boundary Debt)

## Decisions Taken

- **Merge `internal/board` into `dashboard`**:
    - *Rejected*: Keeping `board` as a peer slice.
    - *Reason*: `board` is a DTO-only package serving the UI.
    - *Lean*: Fold DTOs into the `dashboard` slice.

- **Reclassify `internal/cliexec` as a Library**:
    - *Rejected*: Keeping `cliexec` as a `kind: cli` surface.
    - *Reason*: It functions as a technical utility (Runner) for other packages.
    - *Lean*: Move to `libraries[]`.

- **Reclassify `internal/llm` as a Library**:
    - *Rejected*: Keeping `llm` as a peer domain slice.
    - *Reason*: LLM integration is a capability used by others, not a core bounded context.
    - *Lean*: Move to `libraries[]`.

- **Consolidate `internal/server` into `dashboard`**:
    - *Rejected*: Keeping `server` as a top-level slice.
    - *Reason*: It is an `http-surface` specifically hosting the dashboard UI.
    - *Lean*: Nest the server logic under the `dashboard` slice.

- **Group `syncproj` and `triage` under `gitboard`**:
    - *Rejected*: Treating them as independent bounded contexts.
    - *Reason*: They represent internal pipeline stages.
    - *Lean*: Group them as `owns[]` components of the `gitboard` slice.

## Technical Debt and Boundary Violations

| Finding | Smell | Alternatives | Lean |
|:---|:---|:---|:---|
| **Dashboard -> Gitboard** | `internal/server` imports `internal/triage` and `internal/pruneagent` without SliceBindings. | 1. Formalize SliceBindings in catalog. 2. Decouple via interfaces/events. | Approve SliceBindings to allow the dashboard to orchestrate these slices. |
| **Localgit -> Gitboard** | `internal/localgit` imports `internal/cliexec` (part of `gitboard` domain). | 1. Move `cliexec` to a shared library. 2. Accept cross-slice binding. | Reclassify `cliexec` as a library to resolve the dependency. |
| **Pruneagent -> Gitboard** | `internal/pruneagent` imports `internal/cliexec` and `internal/llm`. | 1. Formalize bindings. 2. Move utilities to libraries. | Reclassify `cliexec` and `llm` as libraries to clean the graph. |
| **Remotegit -> Dashboard** | `internal/remotegit` imports `internal/board` (DTO). | 1. Formalize SliceBinding. 2. Move DTO to a shared library. | Approve SliceBinding to `dashboard` once DTOs are consolidated. |
| **Remotegit -> Gitboard** | `internal/remotegit` imports `internal/cliexec`. | 1. Formalize binding. 2. Move `cliexec` to a library. | Reclassify `cliexec` as a library. |
| **Orphaned Observability** | `internal/observability` is only imported by `cmd/gitboard`. | 1. Keep as library. 2. Fold into `cmd/gitboard`. | Move to `libraries/observability`. |
| **High-Coupling Hub** | `cmd/gitboard` imports 11 different packages. | 1. Refactor to thin entry point. 2. Accept as composition root. | Accept as a composition root. |