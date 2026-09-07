# Overview

Gitboard is a local dashboard for monitoring GitHub and GitLab activity (CI, MRs, PRs) and local git state.

## Mission
Provide a unified, local-first interface for managing and triaging repository changes and maintenance tasks.

## Architecture
The system is composed of a CLI orchestrator, a web-based dashboard, a git domain for local/remote operations, and automation agents (`pruneagent`, `triage`) that leverage LLM capabilities to assist in repository maintenance.