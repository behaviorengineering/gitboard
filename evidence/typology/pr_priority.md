# Context: Proposed Architecture Grounding for gitboard

Majordomo's Typology digest is proposing a new structural model for the `gitboard` repository on this context branch. This proposal aims to transition the repository from a loose collection of packages into a structured, functional catalog. 

The primary focus of this change is to consolidate technical "leaf" packages—utilities that serve only one caller—into the main `gitboard` orchestration slice. This includes moving observability and sync logic into the `cmd/gitboard` entrypoint. We are also proposing to group the `internal/server` package under a dedicated `gitboard-http` slice to separate the HTTP delivery surface from the CLI interface.

### Architecture Smell: Undocumented Couplings
While the proposed catalog is mostly complete, there is a significant gap in the formalization of how the HTTP surface interacts with the rest of the system. 

Currently, the `internal/server` package (the HTTP delivery surface) directly imports:
- `internal/dashboard` (the dashboard slice for data aggregation)
- `internal/pruneagent` (the pruneagent slice for branch analysis)
- `internal/triage` (the triage slice for automated analysis)

These are "SliceBindings"—approved allowed couplings from one bounded context to another. Because these imports are not yet in the catalog, the architecture is currently "hollow" in these areas.

### Proposed Approach
Majordomo proposes formalizing these relationships by adding the missing `SliceBindings` to the catalog. 

**Alternatives**:
- **Formalize (Lean)**: Add the bindings to the catalog. This acknowledges the HTTP server's role as a coordinator for these domain slices.
- **Refactor (High Cost)**: Re-engineer the server to remove these direct imports. This would be a significant undertaking and may not be necessary if the current orchestration pattern is acceptable to the product team.

**Recommendation**:
Accept the lean approach: **Formalize the bindings**. This ensures the catalog accurately reflects the observed code topology and allows coverage checks to pass without unnecessary refactoring.