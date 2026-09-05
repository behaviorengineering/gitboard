# Conventions

  ## Development Workflow

  - **Configuration**: User configuration is managed via YAML at `~/.config/gitboard/config.yaml` or via the `GITBOARD_CONFIG` environment variable.
  - **Tooling**: The project utilizes `process-compose` for managing the TUI during development (`make serve`).
  - **CLI/TUI**: The project provides both a direct CLI (`gitboard serve`) and an interactive TUI via `make serve`.
  - **Syncing**: Repository discovery and tracking are handled via the `sync` command, which uses an interactive Charm-based TUI.