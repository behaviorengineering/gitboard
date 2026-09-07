# Cluster Proposal: gitboard

  ## Proposed merges
  - **Merge `internal/observability` into `gitboard` (cmd/gitboard)**: `observability` is a sole importer leaf.
  - **Merge `internal/server` into `gitboard` (cmd/gitboard)**: `server` is a sole importer leaf.
  - **Merge `internal/syncproj` into `gitboard` (cmd/gitboard)**: `syncproj` is a sole importer leaf.
  - **Merge `internal/triage` and `internal/pruneagent` into a `triage` domain**: These represent temporal/agentic stages of the same workflow.
  - **Merge `internal/llm` into `triage`**: `llm` is a capability used by the triage/prune workflow.

  ## Proposed renames
  - `cliexec` -> `cliexec` (Keep as internal component, but move under a domain surface).

  ## Anti-pattern findings
  - **Capability as Slice**: `llm` is currently a peer slice. It is a capability, not a domain pillar. It should be a component within the triage/agent domain.
  - **Temporal/Agentic Slices**: `pruneagent`, `triage`, and `syncproj` are currently peer slices. These represent stages of a process and should be consolidated into a single "automation" or "triage" bounded context.
  - **CLI Surface Misclassification**: `cliexec` is listed as a slice with a surface. Since `cliexec` (internal) has `hasMain: false` and only exports `Runner/Exec`, it is an exec-adapter. It should not be a peer slice; it should be a component owned by the domain it serves (e.g., `gitboard` or `remotegit`).
  - **Standalone Server**: `server` is a peer slice. It is a delivery mechanism (surface) for the `dashboard` or `gitboard` domain, not a standalone domain.

  ## Boundary debt
  | Package | Violation | Mitigation |
  |---------|-----------|------------|
  | `llm` | Capability promoted to domain pillar | Move to `internal/triage/llm` or similar |
  | `cliexec` | Exec-adapter treated as peer slice | Demote to `owns[]` of a domain surface |
  | `server` | Delivery mechanism treated as peer slice | Move under `dashboard` or `gitboard` surfaces |

  ## Rationale
  The current topology is "package-heavy," treating every internal directory as a top-level bounded context. This violates DDD by creating peer slices for capabilities (`llm`), delivery mechanisms (`server`, `cliexec`), and temporal stages (`pruneagent`, `triage`). By merging sole-importer leaves into the main entry point (`gitboard`) and grouping agentic capabilities into a single workflow context, we reduce cognitive load and align the architecture with the actual domain model.