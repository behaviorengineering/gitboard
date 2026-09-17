# Grounding
Gitboard serves as a local coordination layer for developer workflows across GitLab and GitHub. It functions as a centralized dashboard for managing code-change state, providing visibility into CI status, open Merge Requests, and failed job triage directly from the local machine.
### Core Typology and Objectives
The system operates through a set of functional modules designed to bridge remote forge state with local development environments:
Triage and Observability: Monitoring failed jobs and managing the flow of incoming work via `triage` and `observability` modules.
Local/Remote Synchronization: Managing the relationship between local worktrees and remote repository states through `localgit` and `remotegit` components.
Project Synchronization: Orchestrating project-level configurations and state via `syncproj`.
Execution and Interface: Providing command-line interaction through `cliexec` and visual status through the `dashboard` and `board` components.
### Operational Scope
Gitboard is a client-side tool that leverages existing forge CLIs (`gh`, `glab`) and local `git` installations. It does not manage remote deployments or runtime operations; instead, it focuses on the developer's local context, facilitating efficient movement between remote repository updates and local worktree management.