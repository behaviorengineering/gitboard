# Cluster Proposal: gitboard

  ## Proposed merges
  - **Merge `internal/observability` into `cmd/gitboard`**: `observability` is a sole importer (only used by the CLI entrypoint) and serves as a setup utility.
  - **Merge `internal/server` into `cmd/gitboard`**: `server` is a sole importer and represents the delivery mechanism for the CLI/Service.
  - **Merge `internal/syncproj` into `cmd/gitboard`**: `syncproj` is a sole importer and appears to be a specific orchestration task for the main command.
  - **Merge `internal/cliexec` into domain surfaces**: `cliexec` is an exec-adapter (hasMain: false, exports Runner/Exec) and should be treated as a utility for domain packages rather than a peer slice.

  ## Proposed renames
  - *No specific renames proposed at this stage, pending merge confirmation.*

  ## Anti-pattern findings
  - **Capability as Pillar**: `llm` is currently a standalone domain slice. It should be treated as a capability used by other domains (like `triage` or `pruneagent`) rather than a primary bounded context.
  - **CLI as Peer Slice**: `cliexec` is listed as a peer slice but functions as an execution adapter. It should be moved under the domain surfaces that utilize it.
  - **Pipeline/Task as Slice**: `syncproj` and `triage` risk being treated as temporal stages; they should be validated as part of a broader "Project Management" or "Git Operations" domain if they don't represent distinct entities.

  ## Boundary debt
  | Violation | Description |
  | :--- | :--- |
  | Capability Leak | `llm` is a peer to core domains like `localgit` and `remotegit`. |
  | Adapter Proliferation | `cliexec` is elevated to a top-level slice despite being a low-level execution runner. |
  | Fragmented Delivery | `server`, `syncproj`, and `observability` are separated from the primary entrypoint (`cmd/gitboard`) despite having no other callers. |

  ## Rationale
  The current topology is highly fragmented, with many packages acting as "helper" or "adapter" packages that are only consumed by the main CLI. By merging sole-importer packages into the `gitboard` delivery slice, we reduce the cognitive load of the bounded context map. Furthermore, moving capabilities like `llm` and adapters like `cliexec` from "peer slices" to "capabilities/utilities" aligns the architecture with DDD principles, ensuring that the domain model is driven by entities (like `board` or `git`) rather than technical implementation details.