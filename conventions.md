# Conventions

## Development Workflow
- **Configuration**: Use `gitboard init` to generate default configurations.
- **Running**: Use `make serve` for a TUI-driven development experience via `process-compose`, or `gitboard serve` for a direct dashboard run.
- **Syncing**: Use `gitboard sync` to discover and select tracked projects.

## Architectural Integrity
- **Slice Boundaries**: New packages should be assigned to a specific slice in the Typology catalog.
- **Dependency Management**: Cross-slice imports must be explicitly declared as `sliceBindings`. If a package requires a dependency on another slice, verify if it should be a formal binding or if the logic should be moved to a shared leaf.