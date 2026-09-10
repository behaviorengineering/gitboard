# Operator Briefing: Catalog Synchronization

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: journey.md](journey.md) · [Next: pr_priority.md](pr_priority.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

The Typology digest has identified a mismatch between the observed Go imports and the proposed `refined_catalog_yaml`. Specifically, the `gitboard-http` slice (the HTTP delivery surface) is performing significant orchestration by importing several domain slices, but these relationships are not yet formalized in the catalog.

### Priority Decision: Formalize HTTP Orchestration
The `internal/server` package (part of the `gitboard-http` slice) currently imports `internal/dashboard` (dashboard), `internal/pruneagent` (pruneagent), and `internal/triage` (triage). It also has a relationship with the main entrypoint via `cmd/gitboard`.

**The Smell**: These are "ghost" couplings. The code relies on these imports to function, but the architecture manifest (the catalog) doesn't acknowledge them, which will cause coverage check failures.

**Alternatives**:
1. **Formalize the bindings**: Explicitly add these `SliceBindings` to the catalog. This recognizes the HTTP surface as a legitimate orchestrator of these domain slices. (Low cost, high clarity).
2. **Decouple the HTTP surface**: Refactor the server to use interfaces or a different communication pattern to remove these direct imports. (High cost, high architectural purity).

**Recommended Lean**:
**Formalize the bindings.** Majordomo proposes adding the missing `SliceBindings` to the catalog. The current design uses the HTTP server as a high-level coordinator for the dashboard, triage, and pruning logic; the catalog should reflect this reality rather than forcing a refactor.

**Evidence Pointers**:
- See `findings_list` for the specific import paths.
- See `journey_md` for the current status of the `gitboard-http` slice definition.
