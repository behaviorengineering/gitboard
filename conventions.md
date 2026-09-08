# Conventions

## Development Workflow
- **Configuration**: User configuration is managed via `~/.config/gitboard/config.yaml`.
- **Execution**: Use `make serve` for the `process-compose` TUI during development, or `gitboard serve` for a direct dashboard run.
- **Syncing**: Use `gitboard sync` to discover and select tracked projects via the interactive TUI.

## Package Organization
- Domain logic resides in `internal/`.
- Entrypoint logic and technical surfaces are contained within the `gitboard` slice.
- Technical utilities like command execution are abstracted through the `cliexec` slice.