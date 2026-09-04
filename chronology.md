# Chronology

  ## Seed Marker
  - **2026-09-04**: Context branch initialized from SHA `5fca41ab`.

  ## Design Decisions
  - **Domain Consolidation**: Merged `server`, `syncproj`, `observability`, `pruneagent`, and `triage` into the `gitboard` slice.
  - **Git Engine Creation**: Unified `localgit` and `remotegit` into a single `git-engine` context.
  - **Capability Reclassification**: Reclassified `llm` as `gitboard-llm-gateway` to reflect its role as a supporting capability rather than a primary domain slice.
  - **Surface Nesting**: Moved `cliexec` to be a surface component of the `gitboard` slice.