# Architecture

  Gitboard is organized into bounded contexts that separate remote forge interactions, local filesystem operations, and user interfaces.

  ## Bounded Contexts

  ### Core Orchestration (`gitboard`)
  Orchestrates the CLI and integrates core system capabilities. It manages observability, sync processes, and command execution.

  ### Configuration (`config`)
  Provides centralized configuration management for all system components.

  ### Project State (`board`)
  Manages the core state and visibility of project status and branch origins.

  ### Remote Interactions (`remotegit`)
  Manages interactions with remote git hosting providers like GitHub and GitLab.

  ### Local Operations (`localgit`)
  Handles local filesystem git operations and worktree inspections.

  ### User Interface (`dashboard`)
  Serves as the visual interface for monitoring project and branch status via a web-based dashboard.

  ### Maintenance &amp; Intelligence (`pruneagent` &amp; `triage`)
  - **Triage**: Analyzes project state to identify candidates for maintenance.
  - **Pruneagent**: Automates the decision-making process for cleaning up stale branches, utilizing LLM capabilities for intelligent triage.

  ## Boundary Debt
  - `internal/triage` may represent a temporal stage slice rather than a stable domain.
  - `internal/llm` is currently nested under `pruneagent` but may require a dedicated capability slice if shared by other domains in the future.