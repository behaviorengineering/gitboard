# Operator Briefing: Architecture Boundary Decisions

The refinement process has identified several cross-slice imports that lack formal `SliceBinding` declarations. These are not necessarily "bugs," but they are undocumented couplings that the machine cannot validate without your leadership.

### Priority Decision: The "Board" Library Bindings
A significant cluster of findings involves slices (`gitboard`, `pruneagent`, `triage`) importing packages currently owned by the `board` slice (specifically `internal-config` and `internal-llm`). 

**The Situation:** We have reclassified `config` and `llm` as technical libraries within the `board` domain to reduce cognitive load. However, because they are now "cross-slice" imports, they must be explicitly permitted.

**The Smell:** If we don't declare these, the Typology validator will flag the architecture as "drifting" even if the code is correct.

**Alternatives:**
1. **Approve the Bindings (Lean):** Formally add `SliceBinding` entries from `gitboard`, `pruneagent`, and `triage` to `board`. This acknowledges that these slices rely on the `board` slice's infrastructure.
2. **Re-architect Infrastructure (High Cost):** Move `config` and `llm` out of the `board` slice into a global `libraries` slice. This is cleaner long-term but requires significant catalog and code movement now.

**Recommended Lean:** Approve the bindings to `board`. It acknowledges the current reality with minimal friction.

### Priority Decision: The "Gitboard" Utility Bindings
The `localgit`, `pruneagent`, and `remotegit` slices are all importing `internal-cliexec`.

**The Situation:** `cliexec` is an execution adapter owned by the `gitboard` slice. 

**The Smell:** We are treating `gitboard` as a provider of execution utilities to other domain slices.

**Alternatives:**
1. **Approve the Bindings (Lean):** Declare `SliceBinding` from the consumer slices to `gitboard`.
2. **Extract to Shared Library (Medium Cost):** Move `cliexec` to a dedicated `libraries` slice so it isn't a dependency on the main orchestration slice.

**Recommended Lean:** Approve the bindings to `gitboard` for now to maintain the current consolidation strategy.