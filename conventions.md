# Conventions

  ## Development Workflow
  - **Initialization**: Use `make init` or `gitboard init` to set up the environment.
  - **Synchronization**: Use `make sync` or `gitboard sync` to discover and track projects.
  - **Running the Dashboard**:
    - Use `make serve` to run the `process-compose` TUI.
    - Use `gitboard serve` for a direct dashboard run without the TUI.
  - **Configuration**: User configuration is managed via `~/.config/gitboard/config.yaml`. Environment variables (e.g., `GITBOARD_CONFIG`, `GITBOARD_LLM_API_KEY`) can be used for overrides.

  ## Tooling
  - **Forge CLIs**: Requires `glab` and `gh` for authentication and data retrieval.
  - **Process Management**: `process-compose` is used for the TUI development experience.