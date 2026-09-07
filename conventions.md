# Conventions

## Project Structure
- `cmd/`: Entry points for CLI applications.
- `internal/`: Private application code, organized by domain slice.
- `docs/`: Documentation and assets.

## Development Workflow
- Use `make init` to set up the local environment.
- Use `make sync` for interactive repository discovery.
- Use `make serve` to run the development TUI via `process-compose`.