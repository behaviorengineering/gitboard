# Conventions

  ## Development Workflow
  * **Configuration**: User settings are managed via `~/.config/gitboard/config.yaml`.
  * **Environment**: Use `make` commands for standard workflows (e.g., `make sync`, `make serve`).
  * **TUI usage**: When running via `make serve`, use arrow keys for navigation, space to toggle, and `/` to filter.

  ## Code Structure
  * **Internal Packages**: Domain logic is encapsulated within the `internal/` directory, organized by functional slice.
  * **Command Entrypoints**: CLI-specific logic resides in `cmd/`.