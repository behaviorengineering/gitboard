# Cluster Proposal: gitboard

  ## Proposed merges
  - **Merge `internal/observability` into `gitboard` (CLI surface)**: `observability` is a sole importer of the CLI hub.
  - **Merge `internal/server` into `gitboard` (CLI surface)**: `server` is a sole importer and appears to be a delivery mechanism for the CLI.
  - **Merge `internal/syncproj` into `gitboard` (CLI surface)**: `syncproj` is a sole importer.
  - **Merge `internal/cliexec` into `gitboard` (CLI surface)**: `cliexec` is an exec-adapter (hasMain: false) used by multiple domain packages; it should be a utility within the CLI surface or a shared platform leaf, but currently, it acts as a runner for the CLI.

  ## Proposed renames
  - *None identified.*

  ## Anti-pattern findings
  - **Capability as Slice**: `llm` is currently a standalone domain slice. Per DDD rules, LLM capabilities should be part of a domain (e.g., `triage` or `pruneagent`) rather than a peer pillar.
  - **Execution Adapter as Surface**: `cliexec` is listed as a `kind: cli` surface in the catalog, but `package_contracts` shows `hasMain: false`. This is an exec-adapter, not a CLI surface. It should be demoted to `owns[]` or integrated into the CLI surface.
  - **Fragmented Domain**: `localgit` and `remotegit` are highly coupled to the same concerns; they should be evaluated for a unified `git` domain.

  ## Boundary debt
  | Package | Violation | Debt Type |
  |---------|-----------|-----------|
  | `llm` | Capability promoted to domain pillar | Structural |
  | `cliexec` | Exec-adapter misclassified as CLI surface | Catalog/Type |
  | `observability` | Orphaned leaf (sole importer) | Structural |
  | `server` | Orphaned leaf (sole importer) | Structural |

  ## Rationale
  The current topology is "hub-and-spoke" where `cmd/gitboard` acts as a massive orchestrator. To move toward bounded contexts, we must consolidate the delivery-only packages (`server`, `observability`, `syncproj`) into the primary CLI surface. Furthermore, the `llm` package represents a capability rather than a business domain and should be absorbed by the logic that requires it to prevent "capability-driven" architecture.