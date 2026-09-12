# Operator Briefing: Boundary Debt Resolution

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: journey.md](journey.md) · [Next: pr_priority.md](pr_priority.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

  The Typology refine has identified two cross-slice import violations where the `analysis` slice (specifically `internal/pruneagent`) consumes packages without declared bindings in the `refined_catalog_yaml`. 

  Majordomo has already processed the mechanical merges (e.g., `localgit` + `remotegit` into `git`). The following items are not mechanical errors, but normative decisions regarding how strictly you wish to enforce the catalog against the observed code.

  ### 1. Missing Binding: `analysis` $\rightarrow$ `git`
  **Smell**: `internal/pruneagent` imports `internal/localgit`, but no `analysis` $\rightarrow$ `git` binding exists in the catalog.
  - **Option A (Strict)**: Update `refined_catalog_yaml` to include the `analysis` $\rightarrow$ `git` slice binding.
    - *Cost*: Low (one line in YAML).
  - **Option B (Debt)**: Accept this as temporary boundary debt to be resolved in a future refinement cycle.
    - *Cost*: Low (immediate), but increases "dark matter" in the topology.
  - **Recommended Lean**: **Option A**. The import is evidenced and stable; formalizing it now prevents future drift.

  ### 2. Missing Binding: `analysis` $\rightarrow$ `runner`
  **Smell**: `internal/pruneagent` imports `internal/cliexec`, but no `analysis` $\rightarrow$ `runner` binding exists in the catalog.
  - **Option A (Strict)**: Update `refined_catalog_yaml` to include the `analysis` $\rightarrow$ `runner` slice binding.
    - *Cost*: Low (one line in YAML).
  - **Option B (Debt)**: Accept this as temporary boundary debt.
    - *Cost*: Low (immediate), but leaves the `analysis` slice's dependencies unmapped.
  - **Recommended Lean**: **Option A**. Since `runner` is a stable technical slice, declaring this dependency provides immediate clarity for the `analysis` slice's scope.

  **Evidence Pointers**:
  - See `findings_list` for the specific package paths involved (`internal/pruneagent`, `internal/localgit`, `internal/cliexec`).
  - See `refined_catalog_yaml` to identify where the new `sliceBindings` would be appended.
