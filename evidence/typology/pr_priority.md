# Proposed Typology Digest: gitboard

Majordomo's Typology digest proposes a reorganization of the `gitboard` repository to move from a collection of loosely coupled packages toward a cohesive set of bounded contexts. This proposal consolidates technical utilities and delivery surfaces to clarify the product's core purpose: managing git state.

### Summary of Changes
This context branch proposes a new catalog model where:
- **Domain Slices** (`gitboard`, `dashboard`, `localgit`, `remotegit`, `pruneagent`) are clearly defined by their business objectives.
- **Technical Libraries** (`config`, `cliexec`, `llm`, `observability`) are extracted from the domain to serve as shared infrastructure.
- **Delivery Surfaces** (the `dashboard` API and the `gitboard` CLI) are nested within their respective domain slices.

### Key Architectural Smells &amp; Leans

**1. Technical Capabilities vs. Domain Pillars**
Currently, `internal/llm` (LLM integration) and `internal/cliexec` (a shell command runner) are treated as independent domain slices. This causes `localgit` and `pruneagent` to depend on the `gitboard` domain just to execute commands.
* **Lean**: Reclassify `cliexec` and `llm` as **libraries** to decouple the domain logic from technical utilities.

**2. Unmapped Orchestration**
The `dashboard` server (`internal/server`) is importing `triage` and `pruneagent` directly, but these relationships are not yet formalized in the catalog.
* **Lean**: Approve **SliceBindings** (allowed couplings from one slice to another) to allow the dashboard to orchestrate these processes.

**3. Fragmented Data Models**
The `remotegit` slice (responsible for GitHub/GitLab APIs) is importing `board-dto` (a data contract) from the `dashboard` slice.
* **Lean**: Approve a **SliceBinding** from `remotegit` to `dashboard` to formalize this dependency.

### Glossary
* **SliceBinding**: An approved, allowed coupling from one bounded context (slice) to another or to a library.
* **Library**: A technical package group with no specific product objective (e.g., a configuration loader or an execution runner).
* **Slice**: A bounded context representing a specific area of business logic.