# Overview

Gitboard is a local developer dashboard for managing GitHub and GitLab workflows.

## Mission
To provide a centralized, local interface for monitoring CI status, MRs/PRs, and local git state to reduce developer context-switching.

## Architecture Summary
The system is composed of specialized slices:
- **Orchestration**: `gitboard` (CLI) and `exec-utils` (runners).
- **Domain**: `board` (state), `git` (repo interaction), and `config` (centralized settings).
- **Intelligence**: `llm` (AI integration), `triage` (categorization), and `pruneagent` (cleanup).
- **UI**: `dashboard` (web visualization).

Agents should respect the boundaries between these slices and note existing boundary debt regarding `config` and `git` dependencies.