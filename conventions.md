# Conventions

  ## Development Workflow
  * **CLI First**: The core application is a CLI tool; ensure changes are compatible with `gitboard` command execution.
  * **TUI Support**: Use `process-compose` for the interactive TUI experience during development.
  * **Configuration**: User settings are managed via `~/.config/gitboard/config.yaml`.

  ## Architecture Principles
  * **Slice Ownership**: New functionality should be placed within the appropriate bounded context (slice) rather than being added to a generic `internal` folder.
  * **Infrastructure Separation**: Distinguish between domain logic and technical libraries (like `config` and `llm`).