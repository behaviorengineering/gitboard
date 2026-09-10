# Overview

# Mission

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: README.md](README.md) · [Next: architecture.md](architecture.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

  Gitboard provides a unified, local dashboard for developers to monitor their active work across GitHub and GitLab. It bridges the gap between remote forge metadata (CI status, Pull Requests, Merge Requests) and the local development environment (active worktrees, branches, and local checkouts). 

  By integrating LLM-powered triage, automated branch analysis, and tools to investigate divergent local branches—including automated synchronization of nested submodules during pulls—Gitboard helps developers reduce cognitive load during context switching and streamlines the cleanup of stale local development state.

## Architecture

# Architecture

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: mission.md](mission.md) · [Next: conventions.md](conventions.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

> **Teaching story.** Living project architecture for humans and review grounding. Not Typology seed evidence (see `evidence/typology/architecture_brief.md`).

  Gitboard is organized into functional slices that separate remote data acquisition, local filesystem inspection, and user presentation.

  ## Core Slices

  ### Data &amp; State
  * **Board**: The central synchronization layer providing shared JSON data structures used across the system.
  * **Config**: A shared library for application-wide configuration management.

  ### Forge &amp; Git Interaction
  * **RemoteGit**: Interfaces with GitHub and GitLab APIs to fetch repository metadata and remote status.
  * **LocalGit**: Inspects the local filesystem to manage git worktrees and branches, including capabilities to inspect branch synchronization status and perform fast-forward pulls that include recursive submodule updates (`git submodule update --init --recursive`) to ensure nested pins match the new tip.
  * **CliExec**: A low-level utility for executing external system processes and shell commands.

  ### Intelligence &amp; Automation
  * **LLM**: Provides OpenAI-compatible chat completion capabilities for automated analysis.
  * **Triage**: Uses LLM intelligence to perform automated analysis of failed jobs and open items.
  * **PruneAgent**: Analyzes local branch state to identify candidates for removal.

  ### Presentation &amp; Orchestration
  * **Dashboard**: Aggregates local and remote data into a unified view for user interaction.
  * **Server**: An HTTP delivery surface that serves the dashboard web interface.
  * **Gitboard (CLI)**: The primary entrypoint that orchestrates background synchronization and CLI commands.

  ## System Flow
  The system synchronizes remote forge data via `RemoteGit` and local state via `LocalGit`, merging them into the `Board` state. This state is then consumed by the `Dashboard` and served via the `Server` to the user.
