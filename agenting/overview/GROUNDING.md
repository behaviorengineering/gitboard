# Overview

  Gitboard is a local developer dashboard for monitoring GitHub/GitLab activity and local Git worktrees.

  ## Mission
  To centralize CI status, MR/PR tracking, and local worktree state into a single, high-visibility local interface.

  ## Architecture Summary
  The system is composed of specialized slices:
  - **Orchestration**: `gitboard` (CLI, sync, observability).
  - **Data/State**: `board` (status/history) and `config` (settings).
  - **Interfaces**: `dashboard` (web UI) and `gitboard-cli`.
  - **Git/Forge**: `localgit` (worktrees), `remotegit` (APIs), and `syncproj` (remote state).
  - **Intelligence**: `llm` and `triage` (AI-driven analysis).
  - **Automation**: `pruneagent` (branch cleanup).