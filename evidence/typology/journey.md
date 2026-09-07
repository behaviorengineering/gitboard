# Journey: gitboard Typology Refinement

## Status
Refinement complete.

## Decisions
- **Merged Sole Importers**: Moved `internal/observability`, `internal/server`, and `internal/syncproj` into the `gitboard` slice to reduce top-level clutter and align with the primary entry point.
- **Demoted `cliexec`**: Moved `internal/cliexec` from a `kind: cli` surface to an `owns[]` component within `gitboard` because it lacks a `hasMain: true` entry point (it is an adapter, not a delivery tool).
- **Reclassified `server`**: Moved `internal/server` from `surfaces` to `owns` as it is a package implementation, not a user-facing interaction layer.
- **Capability Relocation**: Moved `internal/llm` under `pruneagent` as a component to resolve the "Capability as Slice" anti-pattern.
- **Objective Sanitization**: Replaced all hollow template objectives with concrete business-value statements.

## Technical Debt and Boundary Violations

| Package | Issue | Debt Type |
|---------|-------|-----------|
| `internal/triage` | Potential temporal stage slice; logic may belong to a broader domain entity. | Architectural |
| `internal/llm` | Currently nested under `pruneagent`; may eventually need a dedicated capability slice if shared by more domains. | Architectural |