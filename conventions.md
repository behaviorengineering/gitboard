<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

markdown: |
# Conventions
Contributors interact with the system through a local-first orchestration model. Development and workflow revolve around the following principles:
### Tooling and Environment
The project relies on local forge CLIs to manage state and authentication. Contributors maintain their environment using:
`gh` (GitHub CLI)
`glab` (GitLab CLI)
`git` (Local version control)
### Modular Architecture
The codebase follows a strict internal modularity pattern. Contributions are organized into domain-specific packages within the `internal/` directory:
Observability: Managed via `internal/observability`.
Pruning Logic: Handled by `internal/pruneagent`.
Server Interface: Located in `internal/server`.
Project Discovery: Driven by `internal/syncproj`.
### Workflow
Development is centered on local execution and synchronization. The standard lifecycle involves initializing the local configuration via `gitboard init`, synchronizing state with `gitboard sync`, and serving the dashboard via `gitboard serve` to validate changes in a local environment.
