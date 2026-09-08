# Operator Briefing: Priority Architecture Decisions

Majordomo has completed the Typology refine. The following decisions require human leadership to resolve boundary debt.

### 1. The Utility Reclassification (High Priority)
**Situation:** Several slices (`localgit`, `pruneagent`, `remotegit`) are importing packages currently owned by the `gitboard` domain (`cliexec`, `llm`).
**Smell:** We are treating technical capabilities (execution runners and LLM wrappers) as primary domain pillars. This creates a "gravity well" where everything must depend on the `gitboard` slice.
**Alternatives:**
- **Option A (Formalize):** Add SliceBindings to the catalog. *Cost: Low effort, but preserves high coupling.*
- **Option B (Library Extraction):** Move `cliexec` and `llm` into `libraries[]`. *Cost: Moderate refactor, but cleans the topology.*
**Recommended Lean:** **Option B**. Reclassify `cliexec` and `llm` as libraries to decouple the domain slices from technical implementation details.

### 2. Dashboard Orchestration
**Situation:** The `dashboard` server (`internal/server`) is importing `triage` and `pruneagent` directly.
**Smell:** The dashboard is acting as an orchestrator across multiple bounded contexts without explicit permission in the catalog.
**Alternatives:**
- **Option A (Explicit Binding):** Add SliceBindings for `dashboard -> gitboard` and `dashboard -> pruneagent`. *Cost: Minimal.*
- **Option B (Decouple):** Introduce an event bus or interface layer. *Cost: High.*
**Recommended Lean:** **Option A**. Approve the SliceBindings to acknowledge the dashboard's role in triggering these processes.

### 3. Data Contract Locality
**Situation:** `remotegit` is importing `board-dto` from the `dashboard` slice.
**Smell:** A remote-fetching slice is depending on a UI-specific data contract.
**Alternatives:**
- **Option A (Formalize):** Add SliceBinding `remotegit -> dashboard`. *Cost: Low.*
- **Option B (Shared Library):** Move DTOs to a technical library. *Cost: Moderate.*
**Recommended Lean:** **Option A** (temporarily) until the `dashboard` slice is fully consolidated.