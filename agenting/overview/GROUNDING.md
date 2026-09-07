# Overview

## Mission
Gitboard provides a local dashboard for monitoring GitHub/GitLab activity and managing local git worktrees, utilizing LLMs for automated triage.

## Architecture Summary
The system is partitioned into six primary slices:
1. **gitboard**: Orchestration and domain logic.
2. **config**: Centralized settings.
3. **dashboard**: UI surface.
4. **localgit**: Local filesystem/worktree management.
5. **remotegit**: Remote forge synchronization.
6. **triage**: Agentic LLM workflows.

## Agentic Focus
When performing tasks, respect the boundaries between `localgit` (filesystem) and `remotegit` (API). Be aware of existing boundary debt in the `dashboard` and `triage` slices when proposing changes to dependency graphs.