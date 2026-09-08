# Gitboard

Local code-change board for GitLab and GitHub: CI status, open MRs/PRs, failed job triage, and local checkout / worktree state. Uses `gh`, `glab`, and `git` on your machine.

This documentation represents the context branch for the `gitboard` repository, seeded from the initial discovery phase.

## Installation

1. **Install Forge CLIs**:
   ```bash
   brew install glab
   glab auth login
   gh auth login
   ```

2. **Setup Gitboard**:
   Download the latest release, unpack, and add to your `PATH`.

3. **Initialize**:
   ```bash
   gitboard init
   gitboard sync
   gitboard serve
   ```

## Configuration

User configuration is located at `~/.config/gitboard/config.yaml`. Key sections include:
- `llm`: AI triage settings.
- `ui`: Dashboard refresh intervals and branch filtering.
- `local`: Disk roots for scanning checkouts.
- `sync`: GitHub/GitLab discovery settings.
- `projects`: Curated tracked repositories.