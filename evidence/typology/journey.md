# Journey Notes: gitboard Typology Refinement

## Status
Refinement complete. Applied cluster proposal to consolidate fragmented slices and resolve capability leaks.

## Decisions Taken
1. **Consolidation**: Merged `observability`, `server`, and `syncproj` into the `gitboard` slice as they were sole-importers.
2. **Surface Integration**: Moved `cliexec` from a peer slice to a surface of `gitboard`.
3. **Domain Grouping**: Renamed `localgit` and `remotegit` to a unified `internal/git` domain.
4. **UI Realignment**: Moved `dashboard` to `internal/ui/dashboard` to reflect its role as a projection.
5. **Capability Migration**: Identified `llm` as a capability; planned integration into `triage` and `pruneagent` to resolve the "Capability as Domain Pillar" anti-pattern.

## Technical Debt and Boundary Violations

| Violation | Description |
| :--- | :--- |
| Capability Leak | `llm` is currently a standalone package; requires refactor to be consumed as a capability by `triage`/`pruneagent`. |
| Fragmented Git Domain | `localgit` and `remotegit` paths must be physically moved to `internal/git` to match new typology. |
| Surface/Domain Confusion | `cliexec` is currently a package; needs to be treated strictly as a CLI surface component. |