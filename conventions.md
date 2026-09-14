<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

markdown: |
# Conventions
Development and contributions to Gitboard center on local-first orchestration of forge state. Contributors maintain the project by leveraging existing CLI toolchains—specifically `gh`, `glab`, and `git`—to manage worktrees and local checkouts.
The repository follows a modular domain structure, separating concerns between local Git state management, observability, and remote forge API interaction. Contributions typically involve extending these domain layers or refining the integration between local environments and remote repository metadata.
