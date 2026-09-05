# Journey Notes: gitboard Typology Refinement

## Status
Refined via Cluster Proposal. Moving from fragmented peer-slice model to Bounded Context model.

## Decisions Taken
1. **Consolidated Core**: Merged `observability`, `server`, `syncproj`, and `dashboard` into the `gitboard` domain to reduce cognitive load.
2. **Engine Extraction**: Grouped `localgit` and `remotegit` into `git-engine` to separate git primitives from orchestration.
3. **Surface Unification**: Reclassified `cliexec` from a peer slice to a surface of the `gitboard` domain.
4. **Capability Reclassification**: Moved `llm` from a top-level domain to a capability owned by `triage`.
5. **Worker Promotion**: Moved `pruneagent` into `gitboard` as a specialized worker component.

## Technical Debt and Boundary Violations

| Violation | Description |
| :--- | :--- |
| Capability Leak | `llm` was previously a top-level domain; now nested under `triage`. |
| Surface Fragmentation | `cliexec` and `cmd/gitboard` were split; now unified under `gitboard` surfaces. |
| Task-as-Domain | `pruneagent` and `syncproj` were defined as domains; now treated as components/workers. |