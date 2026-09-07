# Cluster Proposal: gitboard

  ## Proposed merges
  - **Merge `internal/observability` into `gitboard` (CLI surface)**: `observability` is a sole importer of nothing but is only imported by `cmd/gitboard`. It is a utility leaf.
  - **Merge `internal/server` into `gitboard` (CLI surface)**: `server` is a sole importer of nothing but is only imported by `cmd/gitboard`.
  - **Merge `internal/syncproj` into `gitboard` (CLI surface)**: `syncproj` is a sole importer of nothing but is only imported by `cmd/gitboard`.
  - **Merge `internal/cliexec` into `gitboard` (CLI surface)**: `cliexec` is an exec-adapter (hasMain: false, exports Runner/Exec). It should not be a peer slice; it is a utility used by the CLI surface.

  ## Proposed renames
  - *None identified.*

  ## Anti-pattern findings
  - **Capability as Pillar**: `llm` is currently a peer slice. In a DDD-aligned structure, LLM is a capability/utility used by domain logic (like `triage` or `pruneagent`), not a bounded context itself.
  - **CLI as Peer Slice**: `cliexec` is currently a peer slice. It is an execution adapter/utility and should be part of the `gitboard` surface or a shared platform utility.
  - **Fragmented Domain Logic**: The current structure treats every package as a peer slice, leading to a "flat" topology where the primary entry point (`gitboard`) is a hub for dozens of tiny, disconnected slices.

  ## Boundary debt
  | Package | Debt Type | Description |
  |---------|-----------|-------------|
  | `llm` | Domain Violation | Capability (LLM) promoted to a top-level domain slice. |
  | `cliexec` | Structural Violation | Execution adapter promoted to a top-level surface slice. |
  | `observability` | Granularity Debt | Small utility package existing as a standalone slice. |
  | `server` | Granularity Debt | Implementation detail of the CLI/Service entry point existing as a standalone slice. |

  ## Rationale
  The current topology is "package-as-slice," which ignores bounded context boundaries. By applying the merge heuristics, we collapse the "hub-and-spoke" model where `cmd/gitboard` acts as a coordinator for many tiny, single-purpose packages. Moving `cliexec`, `observability`, `server`, and `syncproj` under the `gitboard` slice (or a unified `gitboard` domain) reduces the cognitive load of the slice map and aligns with the reality that these are implementation details of the primary application. `llm` is identified as a capability that should be consumed by domain slices like `triage` rather than standing as a peer.