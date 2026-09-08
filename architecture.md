# Architecture

  > **Teaching story.** Living project architecture for humans and review grounding.

  Gitboard is organized into functional slices that separate domain logic from infrastructure and orchestration.

  ## Bounded Contexts

  ### Core Orchestration (`gitboard`)
  The central hub that orchestrates the CLI application, server, and observability. It owns the execution adapter (`cliexec`) and manages the lifecycle of the dashboard and background tasks.

  ### Project Board &amp; Domain (`board`)
  Manages the visual representation and state of project boards. It acts as a primary provider for configuration and LLM capabilities used by other slices.

  ### User Interface (`dashboard`)
  A web-based interface providing a visual dashboard for interacting with project data and management tasks.

  ### Git Integration Slices
  * **`localgit`**: Interfaces with the local filesystem to inspect and manage git worktrees and repositories.
  * **`remotegit`**: Communicates with remote git providers to fetch and analyze repository state.
  * **`triage`**: Analyzes project state to suggest necessary maintenance and cleanup actions.
  * **`pruneagent`**: Automates the decision-making process for cleaning up stale project branches and resources.

  ## Component Relationships
  The architecture follows a pattern where specialized slices (like `triage` or `pruneagent`) consume data from the `board` (for config/LLM) and `localgit` to perform their duties.
