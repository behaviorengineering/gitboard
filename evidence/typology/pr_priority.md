# Typology Digest: Catalog Grounding for `gitboard`

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: human_intervention.md](human_intervention.md) · [Next: README.md](../../README.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

  Majordomo's Typology digest is proposing a new catalog model for this branch to stabilize the repository's architecture. This proposal moves away from a fragmented 13-package sprawl toward a structured set of domain slices and technical libraries. Specifically, we are proposing to group git-related adapters into a single `git` slice and consolidate analysis-related aggregators into an `analysis` slice.

  ### Proposed Model
  This context catalog proposes formalizing several roles:
  - **`dto/board`**: A technical library for JSON data shapes (formerly `internal/board`).
  - **`runner`**: A technical slice for process execution (formerly `internal/cliexec`).
  - **`git`**: A domain slice consolidating git forge adapters (merging `localgit` and `remotegit`).
  - **`analysis`**: A domain slice consolidating project health aggregators (merging `pruneagent` and `triage`).

  ### Priority Findings &amp; Recommendations

  The following boundary violations were detected between the observed code and the proposed catalog.

  #### 1. Unmapped Dependency: `analysis` $\rightarrow$ `git`
  The `analysis` slice (specifically the `internal/pruneagent` package) imports the `git` slice (`internal/localgit`), but this relationship is not yet declared in the catalog.
  - **Approach A (Formalize)**: Update the catalog to include the `analysis` $\rightarrow$ `git` slice binding. This is low cost and provides immediate architectural clarity.
  - **Approach B (Defer)**: Accept this as temporary boundary debt to be resolved in a future cycle.
  - **The Lean**: **Formalize the binding.** The import is stable and evidenced; declaring it now prevents the topology from becoming "dark matter."

  #### 2. Unmapped Dependency: `analysis` $\rightarrow$ `runner`
  The `analysis` slice (`internal/pruneagent`) also imports the `runner` slice (`internal/cliexec`) without a declared binding.
  - **Approach A (Formalize)**: Add the `analysis` $\rightarrow$ `runner` slice binding to the `refined_catalog_yaml`.
  - **Approach B (Defer)**: Leave the dependency unmapped for now.
  - **The Lean**: **Formalize the binding.** Since `runner` is a stable technical slice, mapping this dependency clarifies the operational scope of the `analysis` slice.
