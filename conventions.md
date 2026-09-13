<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

markdown: |
# Conventions
Development and contribution to Gitboard center on a modular Go architecture. The repository is organized into specialized internal packages that separate domain logic from interface concerns:
Core Logic: Domain-driven packages reside in `internal/`, such as `internal/localgit` for local state and `internal/remotegit` for forge API interactions.
Interfaces: The `internal/server` and `internal/dashboard` packages manage the HTTP and UI layers, while `cmd/gitboard` provides the primary CLI entry point.
Observability: The system integrates OTEL via `internal/observability` to manage error-only inference dumps.
Contributors interact with the system by leveraging local forge CLIs (`gh`, `glab`) and standard `git` workflows to synchronize and serve the local dashboard.
