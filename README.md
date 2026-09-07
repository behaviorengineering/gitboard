# Gitboard

![Gitboard dashboard](docs/banner.webp)

Local code-change board for GitLab and GitHub: CI status, open MRs/PRs, failed job triage, and local checkout / worktree state. Uses `gh`, `glab`, and `git` on your machine.

## Context Branch
This branch serves as the project understanding digest for the `gitboard` repository. It contains the synthesized teaching story and architectural intent derived from the current typology discovery.

## Quick Start

1. **Install Forge CLIs**: `brew install glab`, `glab auth login`, `gh auth login`.
2. **Install Gitboard**: Download the latest release, unpack, and add to `PATH`.
3. **Initialize**:
   ```bash
   gitboard init
   gitboard sync
   gitboard serve
   ```
4. **Access Dashboard**: Open [http://127.0.0.1:1325/](http://127.0.0.1:1325/).

## Configuration
User config is located at `~/.config/gitboard/config.yaml`. Key sections include `llm` (AI triage), `ui` (dashboard settings), `local` (disk roots), `sync` (remote discovery), and `projects` (tracked repositories).