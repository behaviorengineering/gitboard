# Architecture

  Gitboard is organized into bounded contexts that separate core orchestration from configuration management.

  ## Bounded Contexts

  ### gitboard
  The primary application orchestrator. It manages the lifecycle of the dashboard, CLI operations, and domain logic.
  - **Domain Components**: Board management, AI-driven triage, pruning agents, and git operations (local and remote).
  - **Surfaces**: 
    - `gitboard-cli`: Command-line interface.
    - `gitboard-server`: API for the dashboard.
    - `gitboard-ui`: Web-based dashboard interface.

  ### config
  Centralized configuration management for the gitboard ecosystem, handling user settings and project tracking.

  ## Component Mapping
  - **Core Logic**: `./internal/board`, `./internal/triage`, `./internal/pruneagent`
  - **Git Integration**: `./internal/git/local`, `./internal/git/remote`
  - **AI Capabilities**: `./internal/ai`
  - **Interfaces**: `./cmd/gitboard`, `./internal/server`, `./internal/dashboard`

  ## Boundary Debt
  - **Functional Roles**: `triage` and `pruneagent` are currently housed within the `gitboard` domain but represent functional agentic roles that may eventually require independent boundaries.
  - **Capability vs. Service**: `ai` is currently treated as a component within the main domain rather than a standalone service.
  - **Shared Utilities**: `cliexec` acts as a shared utility dependency across multiple domain packages.