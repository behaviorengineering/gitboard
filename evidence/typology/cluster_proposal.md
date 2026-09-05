# Cluster Proposal: gitboard

  ## Proposed merges
  - **Merge `observability` into `gitboard`**: `observability` is a sole importer (only used by `gitboard`).
  - **Merge `server` into `gitboard`**: `server` is a sole importer (only used by `gitboard`).
  - **Merge `syncproj` into `gitboard`**: `syncproj` is a sole importer (only used by `gitboard`).
  - **Merge `cliexec` into `gitboard`**: `cliexec` acts as a CLI execution surface/utility; it should be a surface of the primary domain or part of the main entry point.

  ## Proposed renames
  - `internal/localgit` and `internal/remotegit` -> `internal/git` (Grouped by job family/domain concern).
  - `internal/dashboard` -> `internal/ui/dashboard` (To clarify it is a surface/projection).

  ## Anti-pattern findings
  - **Capability as Domain Pillar**: `llm` is currently a standalone domain slice. It is a capability/service, not a bounded context of the business domain. It should be integrated into the consuming domain (e.g., `triage` or `pruneagent`).
  - **CLI as Peer Slice**: `cliexec` is treated as a peer to domain slices. It is a surface/utility and should be consolidated.
  - **Temporal/Functional Splitting**: `syncproj`, `triage`, and `pruneagent` appear to be functional stages of a pipeline rather than distinct bounded contexts.

  ## Boundary debt
  | Violation | Description |
  | :--- | :--- |
  | Capability Leak | `llm` exists as a top-level domain slice instead of a capability used by others. |
  | Fragmented Git Domain | `localgit` and `remotegit` are split into separate slices despite sharing the same core domain concern. |
  | Surface/Domain Confusion | `cliexec` is a platform/surface utility masquerading as a domain slice. |

  ## Rationale
  The current topology is highly fragmented with many "sole importer" packages that increase cognitive load without providing domain separation. By merging the single-use packages (`server`, `syncproj`, `observability`) into the main `gitboard` context, we reduce the number of slices. Moving `llm` from a pillar to a capability prevents the architecture from being driven by technology rather than business domain.