<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

yaml
markdown: |
# Conventions
Development in `gitboard` follows a modular, domain-driven approach. Contributors maintain the integrity of the system by adhering to the established package boundaries and layer responsibilities.
### Package Structure
The repository organizes logic within the `internal/` directory to enforce encapsulation.
Domain Logic: Core functionality resides in domain-specific packages (e.g., `internal/localgit`, `internal/remotegit`, `internal/pruneagent`).
Infrastructure & Interfaces: Components like `internal/server` provide the interface layer, while `internal/observability` manages telemetry.
Entry Points: Application binaries are located in `cmd/gitboard`.
### Development Workflow
When extending functionality, ensure that new logic is placed within the appropriate domain package to maintain the observed topology. Changes to the core Git interaction logic should be isolated to the `localgit` or `remotegit` modules to preserve the separation between local state management and remote forge API interactions.
