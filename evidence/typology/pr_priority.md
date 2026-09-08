# Proposed Architecture Grounding: gitboard

  Majordomo's Typology digest is proposing a transition from a flat package structure to a context-centric model on this branch. This proposal aims to group specialized logic (like observability and server components) into the primary `gitboard` CLI context and reclassify utility libraries like `cliexec` (an execution adapter) as part of a technical `exec-utils` library.

  ### The Smell
  The current repository treats every package as a peer, which obscures the hierarchy. Furthermore, several critical cross-slice imports exist—such as the `git` domain accessing `board` and `config`, and the `llm` capability accessing `config`—that are not yet formally recorded in the architecture catalog.

  ### Proposed Changes
  - **Context Consolidation**: Grouping `observability`, `server`, and `syncproj` under the `gitboard` slice to reduce fragmentation.
  - **Library Reclassification**: Moving `cliexec` (a runner library, not a main package) into an `exec-utils` slice.
  - **Git Domain Unification**: Merging `localgit` and `remotegit` into a single `git` domain.

  ### Alternatives &amp; Lean
  - **The Formalization Path**: We could refactor the code to remove these imports via interfaces (High Cost), or we can simply formalize the existing imports as explicit SliceBindings (Low Cost).
  - **The Lean**: Majordomo proposes **formalizing the bindings** (e.g., `git -> board`, `gitboard -> exec-utils`, and `llm -> config`) to ground the architecture in its actual observed state.

  *Note: These changes represent a proposed catalog model for grounding; they are not a reorganization that has already been merged into the product repository.*