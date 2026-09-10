# Conventions

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

    ## Development Workflow
    * **Configuration**: User settings are managed via `~/.config/gitboard/config.yaml`.
    * **Environment**: Use `make` commands for standard workflows (e.g., `make sync`, `make serve`).
    * **TUI usage**: When running via `make serve`, use arrow keys for navigation, space to toggle, and `/` to filter.
    * **Git Pulls**: Behind + clean pulls are ff-only and include `git submodule update --init --recursive` to sync nested submodules.

    ## Code Structure
    * **Internal Packages**: Domain logic is encapsulated within the `internal/` directory, organized by functional slice.
    * **Command Entrypoints**: CLI-specific logic resides in `cmd/`.
