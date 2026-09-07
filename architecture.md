# Architecture

  Gitboard is organized into bounded contexts that separate remote forge interaction, local filesystem inspection, and user interface presentation.

  ## Bounded Contexts

  ### Core Orchestration (`gitboard`)
  Provides the primary CLI entry point and orchestrates the application ecosystem. It manages execution adapters (`cliexec`), observability, and project synchronization (`syncproj`).

  ### Project State (`board`)
  Manages the core representation of project status and branch history.

  ### Configuration (`config`)
  Provides centralized configuration management for all system components, including LLM settings and UI preferences.

  ### User Interface (`dashboard`)
  Serves a web-based interface for visualizing project and branch data.

  ### Intelligence (`llm` &amp; `triage`)
  `llm` provides client capabilities for AI interactions, while `triage` uses these capabilities to analyze repository state and provide actionable insights.

  ### Git Interaction (`localgit` &amp; `remotegit`)
  `localgit` interfaces with the local filesystem to inspect worktrees, while `remotegit` manages interactions with GitHub and GitLab APIs.

  ### Automation (`pruneagent`)
  Automates the decision-making process for cleaning up stale branches.

  ## Boundary Debt
  - **LLM Capability**: The `llm` package currently exists as a top-level slice but is a capability that should eventually be an internal utility of domain slices.
  - **Sync Granularity**: `syncproj` is currently owned by the `gitboard` slice but remains a candidate for further consolidation or distinct domain separation.