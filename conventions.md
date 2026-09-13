<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

markdown: |
# Conventions
Development revolves around local state management and forge integration. Contributors interact with the project through standard Git workflows, utilizing `gh` (GitHub CLI) and `glab` (GitLab CLI) to manage remote interactions.
Configuration and environment setup are managed via `gitboard init` and local YAML files in `~/.config/gitboard/`. All changes are tracked through local worktrees and synchronized with the respective forge providers.
