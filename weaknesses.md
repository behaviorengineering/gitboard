# Weaknesses

  * **Catalog-Code Mismatch**: The `gitboard-http` slice (HTTP delivery surface) has several undocumented dependencies on the `dashboard`, `pruneagent`, and `triage` slices.
  * **Unformalized Orchestration**: The `internal/server` package acts as an orchestrator for multiple domain slices, but these `SliceBindings` are not yet fully formalized in the structural catalog.
  * **Entrypoint Coupling**: There is an unrecorded coupling between the `gitboard` CLI entrypoint (`cmd/gitboard`) and the `gitboard-http` slice (`internal/server`).