# Overview

  Gitboard is a local developer dashboard for GitLab and GitHub, focusing on CI status, MR/PR triage, and local worktree management.

  ## Core Architecture
  The system is divided into four primary bounded contexts:
  1. **`gitboard`**: The core orchestration and configuration engine.
  2. **`ui`**: The dashboard projection layer.
  3. **`triage`**: LLM-powered automated triage.
  4. **`pruneagent`**: Automated resource pruning.

  ## Implementation Focus
  Current efforts focus on resolving capability leaks (specifically regarding `llm` integration) and unifying the git domain under `internal/git`.