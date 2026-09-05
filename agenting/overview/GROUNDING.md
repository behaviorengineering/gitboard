# Overview

Gitboard is a local dashboard for managing GitHub and GitLab activity.

## Core Mission
To provide a unified, local interface for monitoring CI, MRs/PRs, and local worktree states.

## Architecture Summary
The system is partitioned into `gitboard` (orchestration), `git-engine` (git primitives), and `triage` (AI-assisted analysis). Development is driven by a CLI and a web-based UI, supported by a `process-compose` TUI for local development.