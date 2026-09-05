# Weaknesses

  ## Known Risks and Gaps
  - **Structural Coupling**: `cliexec` is a shared utility dependency used by multiple domain packages, creating tight coupling within the `gitboard` slice.
  - **Domain Granularity**: Functional roles like `triage` and `pruneagent` are currently grouped under the main `gitboard` domain, which may lead to a bloated bounded context.
  - **Capability Integration**: The `ai` capability is currently implemented as a component within the main domain rather than a decoupled service.