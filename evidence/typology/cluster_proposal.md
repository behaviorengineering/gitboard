# Typology Cluster Proposal: gitboard

## Proposed merges
Majordomo proposes the following merges to consolidate technical companions and reduce slice fragmentation. These are architecture-grounding proposals for the context branch.

*   **`internal/observability` → `cmd/gitboard`**: Typology refine grouped `observability` (role: config) into the entrypoint as it is a sole importer used for bootstrapping.
*   **`internal/server` → `cmd/gitboard`**: Typology refine grouped `server` (role: http_surface) into the entrypoint. While it serves the UI, it currently acts as a specialized surface for the CLI-driven server mode.
*   **`internal/syncproj` → `cmd/gitboard`**: Typology refine grouped `syncproj` (role: aggregator) into the entrypoint as it is a sole importer.

## Proposed renames
To align the catalog with observed roles and resolve "unknown" classifications:

*   **`internal/localgit`** $\rightarrow$ **`internal/localgit` (adapter)**: Typology refine grouped this as an adapter for local git operations.
*   **`internal/pruneagent`** $\rightarrow$ **`internal/pruneagent` (exec_runner)**: Typology refine grouped this as an `exec_runner` based on its use of `cliexec`.
*   **`internal/remotegit`** $\rightarrow$ **`internal/remotegit` (adapter)**: Typology refine grouped this as an adapter for GitHub/GitLab APIs.
*   **`internal/syncproj`** $\rightarrow$ **`internal/syncproj` (aggregator)**: Typology refine grouped this as an aggregator for project discovery.
*   **`internal/triage`** $\rightarrow$ **`internal/triage` (dto)**: Typology refine grouped this as a DTO provider for triage requests/responses.

## Anti-pattern findings
*   **Capability Promotion**: The current draft catalog treats `localgit`, `remotegit`, and `pruneagent` as distinct slices without defined roles. Typology identifies these as technical adapters and runners rather than independent domain pillars.
*   **Unclassified Aggregators**: `syncproj` and `triage` are currently unclassified, which risks them being treated as domain slices rather than technical orchestration layers.

## Boundary debt
*   **Smell: Role Ambiguity in `localgit` and `remotegit`**: These packages are currently "unknown" in the draft but act as critical adapters.
    *   *Alternatives*: 1. Keep as standalone slices (High cost: increases slice sprawl). 2. Explicitly classify as adapters (Lean: provides clarity for consumers).
    *   *Lean*: Explicitly classify as adapters in the next catalog iteration.
*   **Smell: Sole Importer Bloat**: `cmd/gitboard` is becoming a high-degree hub for technical utilities (`observability`, `server`).
    *   *Alternatives*: 1. Move utilities to a `libraries/` slice (Medium cost: requires path refactoring). 2. Accept the entrypoint as a container for these surfaces (Low cost: current state).
    *   *Lean*: Accept the entrypoint as the primary surface container for this small-scale tool.

## Rationale
The proposals prioritize the observed `package_roles` (AST/import evidence) over folder names. By classifying the "unknown" packages as adapters and runners, Majordomo provides a stable topology that explains why `dashboard` and `gitboard` import them. The merges target sole-importer technical packages to reduce the cognitive load of the slice map.