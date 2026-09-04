# Journey Notes: gitboard Typology Refinement

## Status
Refinement complete. Consolidated flat topology into a hierarchical domain model.

## Decisions Taken
1. **Domain Consolidation**: Merged `server`, `syncproj`, `observability`, `pruneagent`, and `triage` into the `gitboard` slice to reduce the hub-and-spoke complexity.
2. **Git Engine Creation**: Unified `localgit` and `remotegit` into `git-engine` to resolve fragmented git logic.
3. **Capability Reclassification**: Renamed `llm` to `gitboard-llm-gateway` and moved it from a peer slice to a supporting capability.
4. **Surface Nesting**: Moved `cliexec` from a top-level slice to a surface component of the `gitboard` slice.
5. **Naming Alignment**: Renamed task-based slices (`pruneagent`, `triage`) to align with the `gitboard` domain prefix.

## Technical Debt and Boundary Violations

| Violation | Description | Mitigation |
| :--- | :--- | :--- |
| Capability Leak | `llm` was previously a primary domain slice. | Refined to `gitboard-llm-gateway` as a capability. |
| Fragmented Git Logic | `localgit` and `remotegit` were split. | Merged into `git-engine`. |
| Surface/Domain Blur | `cliexec` sat at the same hierarchy level as core logic. | Nested under `gitboard` surfaces. |