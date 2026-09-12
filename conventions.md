# Conventions

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

## Development Workflow

- **Configuration**: User settings are managed via `~/.config/gitboard/config.yaml`.
- **Environment**: The system relies on local forge CLIs (`gh`, `glab`) and `git`.
- **Running the App**: 
    - For standard use: `gitboard serve`.
    - For development with hot-reloading: `make serve` (uses `process-compose`).

## Repository Structure

- `cmd/`: CLI entrypoints.
- `internal/`: Core application logic, organized by functional slices.
- `docs/`: Documentation and assets.
- `Makefile`: Primary orchestration for build, sync, and development tasks.
