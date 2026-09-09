# Architecture

> **Teaching story.** Living project architecture for humans and review grounding. Not Typology seed evidence (see `evidence/typology/architecture_brief.md`).

  Gitboard is organized into functional slices that separate remote data acquisition, local filesystem inspection, and user presentation.

  ## Core Slices

  ### Data &amp; State
  * **Board**: The central synchronization layer providing shared JSON data structures used across the system.
  * **Config**: A shared library for application-wide configuration management.

  ### Forge &amp; Git Interaction
  * **RemoteGit**: Interfaces with GitHub and GitLab APIs to fetch repository metadata and remote status.
  * **LocalGit**: Inspects the local filesystem to manage git worktrees and branches.
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
