# Architecture

  Gitboard is organized into functional bounded contexts that separate core domain logic from external interfaces and capabilities.

  ## Bounded Contexts

  ### `gitboard` (Orchestrator)
  The main application entry point. It orchestrates the flow between the configuration, the dashboard, the git engine, and the LLM gateway.
  - **Includes**: Server, sync logic, observability, pruning agents, and triage logic.
  - **Surfaces**: CLI (`cmd/gitboard`).

  ### `git-engine` (Git Operations)
  A unified layer for handling both local and remote git operations, abstracting the complexities of different git environments.

  ### `board` (Domain)
  The core domain logic responsible for managing the state of the developer board.

  ### `dashboard` (Presentation)
  The visual representation layer that monitors and displays the current state of the gitboard.

  ### `config` (Configuration)
  Centralized management of user settings, project tracking, and forge authentication.

  ### `gitboard-llm-gateway` (Capability)
  A specialized capability provider that enables LLM-based operations and triage.

  ## Component Bindings

  - `gitboard` reads from `config`, `dashboard`, `gitboard-llm-gateway`, and `git-engine`.
  - `dashboard` reads from `board`, `config`, and `git-engine`.
  - `git-engine` reads from `board` and `config`.
  - `gitboard-llm-gateway` reads from `config`.