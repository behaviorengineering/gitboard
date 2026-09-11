# Conventions

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

Contributors to Gitboard follow these patterns to maintain modularity and clarity:

## Code Organization
- **Domain Logic**: Business logic and data shapes reside in `internal/`.
- **Entrypoints**: CLI commands and application bootstrapping reside in `cmd/`.
- **Separation of Concerns**: Packages should communicate via defined DTOs (primarily in `internal/board`) to avoid tight coupling between forge adapters and the UI.

## Development Workflow
- **Tooling**: Use `make` for common tasks (`make init`, `make sync`, `make serve`).
- **Testing**: Ensure new features in `internal/` are accompanied by unit tests.
- **Configuration**: Respect the `~/.config/gitboard/config.yaml` structure.

## Documentation
- Updates to architecture or mission must be reflected in the context branch markdown files.
- Use the `typology` tool to validate package boundaries and roles.
