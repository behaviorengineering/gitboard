# Overview

## Mission
Provide a local dashboard for GitHub/GitLab workflow management, including CI status, MR/PR tracking, and local git state.

## Architecture Summary
The system is composed of a central **Dashboard** orchestrator that consumes data from **LocalGit** and **RemoteGit** adapters. Specialized logic for **Triage** and **PruneAgent** utilizes an **LLM** slice for analysis. The **Gitboard** slice serves as the primary entrypoint for both the CLI and the local server.

## Key Slices
- `board`: Shared state structures.
- `dashboard`: Unified view orchestration.
- `localgit` / `remotegit`: External tool adapters.
- `cliexec`: System command execution.
- `llm`: AI-driven analysis.