# Overview

  Gitboard is a local developer tool for monitoring GitHub/GitLab activity and local git state.

  ## Core Mission
  To provide a unified dashboard for CI status, MR/PR management, and job triage, integrated with local worktree information.

  ## Architectural Foundation
  The system is composed of specialized slices:
  - **Infrastructure**: `config`, `cliexec`, `git`.
  - **Domain/Logic**: `board`, `llm`, `pruneagent`, `triage`.
  - **Delivery**: `dashboard` (UI) and `gitboard` (CLI/Server).

  The architecture emphasizes a clear separation between the git infrastructure services and the high-level orchestration performed by the `gitboard` and `dashboard` components.