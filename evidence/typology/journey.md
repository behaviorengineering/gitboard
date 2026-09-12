# Journey: Typology Refinement (gitboard)

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: cluster_proposal.md](cluster_proposal.md) · [Next: human_intervention.md](human_intervention.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

**Status: Open**

## Decisions Taken
- **Merged `localgit` and `remotegit` into `git` slice**: Both serve as adapters for external git data. Rejected keeping them separate to reduce adapter sprawl.
- **Merged `pruneagent` and `triage` into `analysis` slice**: Both function as aggregators for project health/triage. Rejected keeping them separate to consolidate "dark matter" unknowns.
- **Renamed `board` to `dto/board`**: Explicitly signaled its role as a technical data-shape library.
- **Renamed `cliexec` to `runner`**: Aligned with the `exec_runner` role.
- **Verbatim Objective Mapping**: Applied ledger objectives for `git` (using remote forge objective) and `analysis` (using triage objective) to satisfy validation.

## Technical Debt and Boundary Violations

| Smell | Alternatives | Lean |
|-------|--------------|------|
| **Aggregator/Domain Blur**: `dashboard` composes `git` adapters directly. | Move orchestration to a dedicated `orchestrator` slice. | Accept `dashboard` as an aggregator and accept the coupling. |
| **Role Ambiguity**: High number of `unknown` roles in `git` and `analysis` components. | Perform deep AST scan to assign specific roles. | Accept debt in this pass; flag for next Typology cycle. |
| **Missing SliceBinding (Git)**: `internal/pruneagent` (analysis) imports `internal/localgit` (git) without a declared binding. | Update `refined_catalog_yaml` to include the `analysis -> git` binding. | Accept as temporary boundary debt; update catalog in next refinement. |
| **Missing SliceBinding (Runner)**: `internal/pruneagent` (analysis) imports `internal/cliexec` (runner) without a declared binding. | Update `refined_catalog_yaml` to include the `analysis -> runner` binding. | Accept as temporary boundary debt; update catalog in next refinement. |
