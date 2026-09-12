# Weaknesses

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: conventions.md](conventions.md) · [Next: chronology.md](chronology.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

The following items represent known boundary debt or architectural smells identified during typology refinement.

- **Analysis Boundary Debt (Git)**: The `analysis` slice (specifically `internal/pruneagent`) imports `internal/localgit` (part of the `git` slice) without a formally declared `SliceBinding` in the catalog.
- **Analysis Boundary Debt (Runner)**: The `analysis` slice (specifically `internal/pruneagent`) imports `internal/cliexec` (part of the `runner` slice) without a formally declared `SliceBinding` in the catalog.
- **Aggregator Coupling**: The `dashboard` slice currently composes `git` adapters directly. This creates a tight coupling between the presentation layer and forge adapters; a future refactor might introduce a dedicated `orchestrator` slice.
- **Role Ambiguity**: Several components within the `git` and `analysis` slices currently lack specific role assignments in the typology, requiring further AST-level investigation.
