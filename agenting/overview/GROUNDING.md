# Overview

  ## Mission
  To provide a local, unified dashboard for monitoring multi-forge repository state and automating maintenance.

  ## Architecture Summary
  The system is composed of specialized slices:
  * **Orchestration**: `gitboard`
  * **Domain/State**: `board`
  * **Interface**: `dashboard`
  * **Git Operations**: `localgit`, `remotegit`, `triage`, `pruneagent`

  Agents should respect these boundaries and note existing coupling debt between the orchestration slice and the domain/infrastructure libraries.