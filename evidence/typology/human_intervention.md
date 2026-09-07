# Operator Intervention Required

  The following architecture findings represent structural drift. Do **NOT** attempt to resolve these by inventing `sliceBindings` in the catalog unless the dependency is intentional and permanent.

  ## Priority Decisions

  ### 1. Binding Approvals (High Priority)
  Decide if the following cross-slice imports should be formally declared in the catalog or if the code should be refactored to remove them:
  - **Config Access**: `dashboard`, `git`, and `triage` all import `config`. Is `config` a shared platform leaf, or should these slices receive config via dependency injection?
  - **Git/Board Coupling**: `remotegit` imports `board`. Is this a valid dependency, or should `board` be a lower-level utility?
  - **Execution Adapter**: `git`, `gitboard`, and `git` (local/remote) all import `cliexec`. Confirm `cliexec` is a valid shared utility.

  ### 2. Capability vs. Domain (Medium Priority)
  - **LLM Integration**: `pruneagent` and `triage` import `llm` (owned by `gitboard`). Decide if `llm` should remain a capability within `gitboard` or if the bindings should be formalized to treat `llm` as a shared service.

  ### 3. Refactor vs. Accept (Low Priority)
  - **Pruneagent Dependencies**: `pruneagent` imports `localgit` (git) and `llm` (gitboard). Determine if `pruneagent` is becoming a "god service" or if these are legitimate domain requirements.

  **Evidence Pointers:**
  - See `architecture_md` for the intended Bounded-context map.
  - See `findings_list` for the specific import violations.