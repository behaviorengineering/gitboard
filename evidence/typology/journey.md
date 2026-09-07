# Journey: Gitboard Typology Refinement

  ## Status
  Open (Pending Architecture Decision)

  ## Decisions
  - **Consolidated Delivery Slices**: Merged `observability`, `server`, and `syncproj` into the `gitboard` CLI surface as they are sole importers and delivery-only.
  - **Demoted cliexec**: Moved `cliexec` from `kind: cli` to `owns[]` because `hasMain: false` identifies it as an infrastructure adapter, not a user-facing surface.
  - **Unified Git Domain**: Grouped `localgit` and `remotegit` under a single `git` slice to resolve fragmentation.
  - **LLM Capability Handling**: Instead of a standalone domain slice, `llm` is treated as a capability to be absorbed by `triage` and `pruneagent`.

  ## Technical Debt and Boundary Violations

  | Finding | Package/Slice | Violation | Debt Type |
  |---------|---------------|-----------|-----------|
  | 1 | `dashboard` | Missing SliceBinding to `config` | Boundary |
  | 2 | `git` | Missing SliceBinding to `board` | Boundary |
  | 3 | `git` | Missing SliceBinding to `cliexec` (localgit) | Boundary |
  | 4 | `git` | Missing SliceBinding to `cliexec` (remotegit) | Boundary |
  | 5 | `git` | Missing SliceBinding to `config` | Boundary |
  | 6 | `gitboard` | Missing SliceBinding to `cliexec` | Boundary |
  | 7 | `pruneagent` | Missing SliceBinding to `git` | Boundary |
  | 8 | `pruneagent` | Missing SliceBinding to `gitboard` (llm) | Boundary |
  | 9 | `triage` | Missing SliceBinding to `config` | Boundary |
  | 10 | `triage` | Missing SliceBinding to `gitboard` (llm) | Boundary |