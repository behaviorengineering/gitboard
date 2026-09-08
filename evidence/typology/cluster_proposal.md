# Cluster Proposal: gitboard

  ## Proposed merges

  ### Merge `observability`, `server`, and `syncproj` into `gitboard`
  The packages `internal/observability`, `internal/server`, and `internal/syncproj` are currently isolated leaves or near-leaves with only `cmd/gitboard` as a primary consumer. 
  - **Why this grouping:** These packages appear to be implementation details or orchestration logic specifically for the `gitboard` CLI tool rather than independent domain services.
  - **What was rejected:** Keeping them as standalone slices. This would create a fragmented architecture where the "app" is scattered across many tiny, highly-coupled packages.
  - **Cost of alternative:** High maintenance overhead and complex dependency management for a single-binary tool.
  - **The lean:** Consolidate these into the `gitboard` slice to simplify the bounded context.

  ### Merge `cliexec` into `gitboard` (as a library/utility)
  `internal/cliexec` is an exec-adapter (hasMain: false) used by several packages to run commands.
  - **Why this grouping:** It is a technical utility for command execution, not a domain-driven CLI surface.
  - **What was rejected:** Maintaining `cliexec` as a peer `kind: cli` slice.
  - **Cost of alternative:** Violates the rule against creating standalone CLI slices for execution adapters.
  - **The lean:** Move `cliexec` under the `gitboard` slice or a technical `libraries` slice to reflect its role as a helper.

  ## Proposed renames

  ### Rename `internal/config` to `libraries/config`
  - **Why:** The package is a domain-free platform utility (config) used by almost every slice.
  - **The lean:** Categorize as a technical library rather than a domain slice to clarify it provides infrastructure, not business logic.

  ## Anti-pattern findings

  ### Non-CLI CLI Surface (`cliexec`)
  - **Finding:** The `cliexec` slice is declared as `kind: cli`, but `package_contracts` shows `hasMain: false`.
  - **Violation:** `cliexec` is an exec-adapter, not a CLI surface. It should be demoted from `kind: cli` and placed under `owns[]` or a library slice.

  ### Capability-as-Domain (`llm`)
  - **Finding:** `internal/llm` is treated as a standalone domain slice.
  - **Violation:** LLM integration is a capability/tooling concern. It should likely be a library or a component within a higher-level domain (like `triage`) rather than a peer domain pillar.

  ## Boundary debt

  | Smell | Alternatives | Lean |
  | :--- | :--- | :--- |
  | **Fragmented Orchestration:** `server`, `syncproj`, and `observability` are split from the main entrypoint. | 1. Keep as separate slices (High cost/complexity). 2. Merge into `gitboard` (Lean). | Merge into `gitboard` to unify the application logic. |
  | **Technical Utility as Domain:** `config` and `cliexec` are positioned as peer domain slices. | 1. Maintain as peer slices (High cognitive load). 2. Move to `libraries[]` (Lean). | Reclassify as technical libraries to distinguish infra from domain. |

  ## Rationale

  The current inventory reflects a "package-per-folder" structure that has not yet been mapped to a cohesive bounded context. The high out-degree of `cmd/gitboard` suggests it is the actual application, while the other packages are its constituent parts. The proposal moves the repository from a collection of loosely related packages toward a structured architecture where domain logic (like `board` or `triage`) is separated from the orchestration and technical utilities that support the `gitboard` CLI.