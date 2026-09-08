# Operator Briefing: Architecture Boundary Decisions

  Majordomo has completed the initial Typology refine. While the structural reorganization is proposed, several cross-slice imports exist that are not yet formalized in the catalog. You must decide whether to approve these couplings as explicit SliceBindings or refactor the code to decouple them.

  ### Priority Decisions

  #### 1. Git Domain Couplings
  The `git` slice (specifically `internal-remotegit`) currently imports from `board` and `config`.
  - **Smell**: Unrecorded cross-slice dependency.
  - **Alternatives**: 
    - **Approach A (Formalize)**: Add SliceBindings for `git -> board` and `git -> config`. *Cost: Low (Catalog change only).*
    - **Approach B (Decouple)**: Refactor `git` to use interfaces provided by the caller. *Cost: Medium (Code change).*
  - **Recommended Lean**: **Approach A**. Formally record these bindings in the catalog to ground the architecture.

  #### 2. CLI Execution Coupling
  The `gitboard` entry point (`cmd-gitboard`) imports from the `exec-utils` library (`internal-cliexec`).
  - **Smell**: The primary CLI is using a utility library without an explicit binding.
  - **Alternatives**:
    - **Approach A (Formalize)**: Add SliceBinding for `gitboard -> exec-utils`. *Cost: Low.*
    - **Approach B (Consolidate)**: Move the `cliexec` logic directly into the `gitboard` slice. *Cost: Medium.*
  - **Recommended Lean**: **Approach A**. Maintain `exec-utils` as a separate technical library but record the binding.

  #### 3. LLM Configuration Access
  The `llm` slice (`internal-llm`) imports from `config`.
  - **Smell**: Capability-based slice accessing platform configuration.
  - **Alternatives**:
    - **Approach A (Formalize)**: Add SliceBinding for `llm -> config`. *Cost: Low.*
    - **Approach B (Inject)**: Pass configuration values into LLM functions rather than importing the package. *Cost: Medium.*
  - **Recommended Lean**: **Approach A**. Accept the binding to allow the LLM capability to be self-contained regarding its provider settings.

  *For full context on the reorganization, see `journey_md`.*