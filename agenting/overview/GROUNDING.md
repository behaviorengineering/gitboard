# Overview

Gitboard is a local dashboard for GitHub and GitLab that integrates remote forge data with local git worktree state.

## Mission
To provide a unified view of repository health, CI status, and active work (MRs/PRs/checkouts) to streamline developer triage and maintenance.

## Architecture Summary
The system is composed of domain slices (Board, Config, LocalGit, Remotegit, Triage, LLM, PruneAgent) orchestrated via a CLI (`gitboard`) and presented through a web UI (`dashboard`). The architecture prioritizes the separation of remote provider interactions from local filesystem operations.