# Conventions

  ## Development Workflow
  - **Tooling**: Use `make` for common tasks such as `make init`, `make sync`, and `make serve`.
  - **Service Management**: `make serve` utilizes `process-compose` to run a TUI, while `gitboard serve` provides a direct dashboard run.
  - **Configuration**: User settings are managed via `config.yaml`. Environment variables (e.g., `GITBOARD_CONFIG`, `GITBOARD_LLM_API_KEY`) are supported for overrides.

  ## Project Structure
  - `cmd/`: Application entry points.
  - `internal/`: Core logic, organized by bounded context (e.g., `config`, `dashboard`, `remotegit`).
  - `docs/`: Documentation and assets.