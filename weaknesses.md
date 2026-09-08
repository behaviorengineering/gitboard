# Weaknesses

* **High-Coupling Entry Point:** `cmd/gitboard` acts as a massive composition root with 11 outgoing imports, risking a "god package" pattern.
* **Unsanctioned Orchestration:** The `dashboard` server performs cross-slice imports into `triage` and `pruneagent` without explicit SliceBindings in the catalog.
* **Leaky Data Contracts:** `remotegit` depends on `board-dto` within the `dashboard` slice, creating a dependency on UI-centric data structures.
* **Capability-as-Domain Pillar:** `internal/llm` and `internal/cliexec` are currently treated as domain slices rather than technical libraries, causing unnecessary cross-slice coupling.