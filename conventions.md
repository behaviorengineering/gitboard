<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

markdown: |
# Conventions
Contributors manage their local development lifecycle through a CLI-first workflow. The process centers on synchronizing local state with forge providers using `glab` and `gh`.
### Workflow Journey
Initialization: New environments are established via `gitboard init` to configure local settings.
Synchronization: Workflows rely on `gitboard sync` to pull the latest status from GitLab or GitHub.
Observation: Active triage and monitoring occur through the local `gitboard serve` instance, providing a dashboard for CI status, open MRs, and failed job management.
This approach prioritizes local-first visibility into remote forge activity without requiring heavy deployment or runtime operations.
