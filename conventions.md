# Conventions

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture.md](architecture.md) · [Next: weaknesses.md](weaknesses.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

Contributors to `gitboard` work within a local-first development workflow that integrates directly with standard forge CLI tools.
### Environment and Tooling
Development and testing rely on the presence of `gh` (GitHub CLI), `glab` (GitLab CLI), and standard `git` installations. Because `gitboard` functions as a local dashboard for CI status, MRs/PRs, and worktree state, all contributions must ensure compatibility with these underlying command-line interfaces.
### Development Workflow
Installation: For local development, use `go install github.com/behaviorengineering/gitboard/cmd/gitboard@latest` to ensure you are working with the most recent local build.
Configuration: Initial setup is managed via `gitboard init`, which populates the local configuration at `~/.config/gitboard/config.yaml`. Changes to the LLM or provider settings should be reflected in this configuration.
Testing: Since the tool interacts with local worktrees and forge-specific metadata, testing should be performed against active local repositories to validate the aggregation of project views and triage capabilities.
