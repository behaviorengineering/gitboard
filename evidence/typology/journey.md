# Journey: Typology Refinement - gitboard

## Status
Refinement complete. Applied cluster proposal to collapse utility slices into the primary application slice.

## Decisions
- **Collapsed Utility Slices**: `observability`, `server`, `syncproj`, and `cliexec` were moved under the `gitboard` slice to reduce topology fragmentation.
- **Demoted `cliexec`**: Moved from a `kind: cli` surface to `owns[]` within `gitboard` because it is an execution adapter (hasMain: false).
- **Capability Alignment**: Identified `llm` as a capability used by `triage` and `pruneagent` rather than a standalone domain pillar.
- **Objective Hardening**: Replaced all hollow template objectives with concrete business outcomes.

## Technical Debt and Boundary Violations

| Package | Debt Type | Description |
|---------|-----------|-------------|
| `llm` | Domain Violation | Capability (LLM) is still a top-level slice; should eventually be an internal utility of domain slices. |
| `syncproj` | Granularity Debt | Moved to `gitboard` owns[], but remains a candidate for further consolidation or distinct domain separation. |