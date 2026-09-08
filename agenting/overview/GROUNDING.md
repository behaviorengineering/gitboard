# Overview

## Mission
Gitboard is a local dashboard for managing GitHub/GitLab repository lifecycles, combining remote forge data with local Git state for improved triage and environment hygiene.

## Architecture Summary
The system is composed of five primary bounded contexts:
1. **Gitboard Core:** Discovery and triage orchestration.
2. **Dashboard:** UI and API surface.
3. **LocalGit:** Local filesystem and worktree management.
4. **RemoteGit:** Forge API integration.
5. **PruneAgent:** Stale branch/worktree investigation.

Technical capabilities like **Config** and **LLM** are treated as shared libraries.