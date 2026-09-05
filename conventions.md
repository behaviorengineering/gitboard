# Conventions

  ## Development Workflow
  - **Configuration**: User configuration is managed via `~/.config/gitboard/config.yaml`.
  - **Tooling**: The project relies on `gh`, `glab`, and `git` for forge and version control interactions.
  - **Service Management**: Use `make serve` for the `process-compose` TUI during development and `make serve-down` to stop the process.

  ## Package Organization
  - Domain logic is contained within `internal/`.
  - CLI entry points are located in `cmd/`.
  - Structural changes to the package topology should be reflected in the Typology manifest.