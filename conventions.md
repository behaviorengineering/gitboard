# Conventions

## Development Workflow
- **Configuration**: Use `gitboard init` to generate default configurations. Avoid manual edits to `~/.config/gitboard/config.yaml` without referencing `config.example.yaml`.
- **Running the Service**: Use `make serve` for a development environment with a process-compose TUI, or `gitboard serve` for a standard direct dashboard run.
- **Syncing**: Use `gitboard sync` to discover new repositories and update the tracked project list.

## Code Structure
- Logic is contained within the `internal/` directory.
- Command-line entry points reside in `cmd/`.
- Documentation and assets are located in `docs/`.