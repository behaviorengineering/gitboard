# Conventions

  ## Development Workflow

  - **Architecture Alignment**: All new packages must be mapped to a defined slice in the Typology catalog.
  - **Documentation**: Every slice must maintain its own `docs/develop/` sub-directory containing `overview.md` and `components.md` files.
  - **Boundary Integrity**: Avoid cross-slice coupling that is not explicitly declared in the architecture manifest.

  ## Tooling

  - Use `make` for common tasks (init, sync, serve).
  - `process-compose` is used for the TUI-based development environment.