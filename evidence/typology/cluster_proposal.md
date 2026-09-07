# Cluster Proposal: gitboard

  ## Proposed merges
  - **Merge `internal/observability` into `gitboard` (cmd/gitboard)**: `observability` is a sole importer (only used by `gitboard`).
  - **Merge `internal/server` into `gitboard` (cmd/gitboard)**: `server` is a sole importer (only used by `gitboard`).
  - **Merge `internal/syncproj` into `gitboard` (cmd/gitboard)**: `syncproj` is a sole importer (only used by `gitboard`).
  - **Merge `internal/cliexec` into `gitboard` (cmd/gitboard) or relevant domain**: `cliexec` is an exec-adapter (hasMain: false, exports Runner/Exec) and should not be a peer slice. It should be moved under the domain it serves or the CLI surface.

  ## Proposed renames
  - *No renames proposed at this stage.*

  ## Anti-pattern findings
  - **Capability as Slice**: `llm` is a capability, not a domain pillar. It should be a component within a domain (e.g., `triage` or `pruneagent`) rather than a top-level slice.
  - **Standalone CLI/Exec Slice**: `cliexec` is currently defined as a standalone slice. Per instructions, `cliexec` (an exec-adapter) must be demoted from `kind: cli` to `owns[]` or a surface of a domain.
  - **Temporal/Functional Slices**: `triage` and `pruneagent` appear to be functional stages. While they have domain logic, they risk becoming temporal pipeline slices if not anchored to a core entity.

  ## Boundary debt
  | Package | Issue | Debt Type |
  |---------|-------|-----------|
  | `internal/cliexec` | Defined as a standalone `kind: cli` slice despite being an exec-adapter. | Structural |
  | `internal/llm` | Capability-based slice rather than domain-based. | Architectural |
  | `internal/triage` | Potential temporal stage slice. | Architectural |

  ## Rationale
  The current topology treats several utility and capability packages (`cliexec`, `llm`, `observability`, `server`) as peer bounded contexts. Applying the merge heuristics reduces the cognitive load by moving sole-importer packages into the primary entry point (`gitboard`) and demoting infrastructure/exec-adapters to internal components of the domains they support. This aligns the repository with DDD principles by centering the architecture on domain entities rather than technical capabilities or execution patterns.