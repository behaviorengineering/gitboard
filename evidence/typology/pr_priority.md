# Architecture Alignment: Formalizing Slice Bindings

This PR addresses undocumented dependencies discovered during the Typology refinement of the `gitboard` repository. 

### The Situation
We have successfully reorganized the repository to move away from a "package-per-folder" structure toward a bounded-context model. We consolidated orchestration logic into the `gitboard` slice and reclassified infrastructure like `config` and `llm` as technical libraries owned by the `board` slice. 

However, the import graph shows that several other slices rely on these libraries and utilities. Currently, these connections are "invisible" to our architecture validator.

### The Smell
We have "ghost couplings." For example, the `triage` slice (which analyzes project state) imports `internal-config` (a technical library), but there is no `SliceBinding` (an approved allowed coupling from a bounded context to another slice or to a library) to permit this. Without these bindings, we cannot distinguish between intentional architecture and accidental spaghetti code.

### Alternatives
1. **Formalize Bindings (Lean):** We add `SliceBinding` entries to the catalog. This tells the system: "It is intentional that `pruneagent` uses the `llm` library inside `board`."
2. **Decouple via Global Libraries (High Cost):** We move all infrastructure into a completely separate `libraries` slice. While this is a "purer" architecture, it adds significant complexity to the catalog and requires moving multiple packages.

### The Lean
We are proceeding with **Option 1**. We will add the necessary `SliceBinding` entries to the catalog to authorize the existing imports. This allows us to maintain our current consolidation of orchestration logic while satisfying the architecture validator.

**Key Packages Involved:**
- `board` (Domain slice owning `internal-config` and `internal-llm`)
- `gitboard` (Orchestration slice owning `internal-cliexec`)
- `triage`, `pruneagent`, `remotegit`, `localgit` (Consumer slices)