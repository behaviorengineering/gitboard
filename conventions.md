# Conventions

  ## Development Workflow
  - **Configuration**: User configuration is managed via `~/.config/gitboard/config.yaml`.
  - **TUI Usage**: The system utilizes `process-compose` for TUI-based services and `make` commands for orchestration.
  - **CLI Integration**: The tool relies on existing authenticated forge CLIs (`gh`, `glab`) and local `git` installations.

  ## Architecture Maintenance
  - All new packages must be mapped to a bounded context in the Typology catalog.
  - Cross-slice imports must be declared in the `sliceBindings` of the typology manifest.