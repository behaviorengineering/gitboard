# Conventions

## Development Workflow
- Use `make init` and `make sync` to set up the local environment.
- Use `make serve` for a process-compose TUI development experience, which enables automatic restarts on code changes.
- Use `gitboard serve` for a standard direct dashboard run.

## Package Structure
- Domain logic resides in `internal/` sub-packages organized by bounded context.
- CLI-specific orchestration and server management are contained within the `gitboard` slice.
- Utility logic (like `cliexec`) is treated as an internal adapter rather than a standalone domain slice.