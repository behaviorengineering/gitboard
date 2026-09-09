# Overview

  Gitboard is a local-first dashboard for monitoring GitHub/GitLab activity and local git state.

  ## Mission
  To provide a unified view of remote forge status and local worktrees, enhanced by LLM-driven triage and automated branch management.

  ## Architecture Summary
  The system operates through a series of specialized slices:
  - **Data Acquisition**: `RemoteGit` (API) and `LocalGit` (Filesystem).
  - **Intelligence**: `LLM` and `Triage` for automated analysis.
  - **State**: `Board` for cross-component synchronization.
  - **Interface**: `Dashboard` and `Server` for user presentation, orchestrated by the `Gitboard` CLI.