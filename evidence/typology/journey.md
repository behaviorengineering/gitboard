# Journey: Typology Refinement (gitboard)

## Status
Refinement complete. The topology has been collapsed from a fragmented set of helper slices into a cohesive delivery-oriented structure.

## Decisions
- **Consolidated Delivery**: Merged `observability`, `server`, and `syncproj` into the `gitboard` slice as they are sole-importers of the main CLI entrypoint.
- **Adapter Reclassification**: Removed `cliexec` as a peer slice. It is now treated as an internal utility/adapter used by domain packages (not explicitly listed as a surface to avoid "adapter as slice" anti-pattern).
- **Capability Alignment**: Re-contextualized `llm` as a capability used by `triage` and `pruneagent` rather than a primary domain pillar.
- **Surface Mapping**: Moved `dashboard` to `kind: ui` as it represents a user-facing interaction artifact.

## Technical Debt and Boundary Violations

| Violation | Description |
| :--- | :--- |
| Capability Leak | `llm` remains a top-level slice; while it is a capability, its current package structure warrants a slice for now. |
| Fragmented Logic | `cliexec` logic is currently distributed; moving it fully into domain-specific adapters is a future task. |
| Documentation Gaps | Multiple packages lack `docs/develop/` overview and component markdown files. |