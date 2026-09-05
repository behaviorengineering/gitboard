# Conventions

## Development Workflow
- **Tooling**: Uses `make` for orchestration (`make init`, `make sync`, `make serve`).
- **TUI**: The `make serve` command utilizes `process-compose` to run a TUI dashboard.
- **Configuration**: User configuration is managed via YAML at `~/.config/gitboard/config.yaml`.

## Repository Structure
- `cmd/`: Entry points for CLI tools.
- `internal/`: Private library code, organized by domain slice.
- `docs/`: Documentation and assets.