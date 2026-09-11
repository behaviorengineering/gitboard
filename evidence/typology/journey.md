# Typology Refinement Journey: gitboard

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: cluster_proposal.md](cluster_proposal.md) · [Next: human_intervention.md](human_intervention.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

## Status
**Refined.** The catalog has been updated to resolve unknown roles and validation errors using evidence from the ledger and capability constraints.

## Decisions Taken
- **Resolved Unknown Roles**: Majordomo assigned roles to `internal/localgit` (inspector/localgit), `internal/pruneagent` (pruneagent), `internal/remotegit` (adapter), `internal/syncproj` (aggregator), and `internal/triage` (aggregator) based on `package_contracts` and `slice_objective_ledger_yaml`.
- **Fixed Validation Errors**: Corrected `gitboard-cli` and `server-ui` surface kinds to `cli` and `ui` respectively to satisfy schema requirements.
- **Objective Injection**: Verbatim objectives from `slice_objective_ledger_yaml` were injected into all slices to replace hollow or missing strings.
- **Rejected Merges**: Majordomo rejected the proposal to merge `observability`, `server`, or `syncproj` into the `gitboard` entrypoint to maintain modularity and distinct delivery concerns.

## Technical Debt and Boundary Violations

| Smell | Alternatives | Lean |
|-------|--------------|------|
| **Unknown Role Debt** (Resolved) | 1. Manual assignment. 2. Leave as unknown. | Majordomo assigned roles based on observed contracts to enable automated enforcement. |
| **Documentation Gap** | 1. Bulk generate stubs. 2. Ignore until feature-complete. | Prioritize documentation for `board` and `cliexec` as core technical dependencies. |
| **Ambiguous Naming** | 1. Rename `remotegit` to `forge_adapter`. 2. Keep `remotegit`. | Majordomo proposes renaming `remotegit` to `forge_adapter` to clarify its role as an adapter. |
| **Ambiguous Naming** | 1. Rename `localgit` to `git_inspector`. 2. Keep `localgit`. | Majordomo proposes renaming `localgit` to `git_inspector` to reflect its role in inspecting worktrees. |
