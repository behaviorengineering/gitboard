# Gitboard

  ![Gitboard dashboard](docs/banner.webp)

  Local code-change board for GitLab and GitHub: CI status, open MRs/PRs, failed job triage, and local checkout / worktree state. Uses `gh`, `glab`, and `git` on your machine.

  ## Context Branch
  This README is part of a context branch holding project understanding for the `gitboard` repository.

  ## Installation (Binary)
  1. Install and log in to forge CLIs (`glab`, `gh`).
  2. Download the `gitboard` archive, unpack, and add to `PATH`.
  3. Run `gitboard init`, `gitboard sync`, and `gitboard serve`.

  ## From Source
  ```bash
  make init
  make sync
  make serve
  ```

  ## Configuration
  User config lives at `~/.config/gitboard/config.yaml`. Key sections include `llm` (AI triage), `ui` (dashboard settings), `local` (disk roots), `sync` (remote discovery), and `projects` (tracked repos).