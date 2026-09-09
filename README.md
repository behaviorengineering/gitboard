# Gitboard

  ![Gitboard dashboard](docs/banner.webp)

  Local code-change board for GitLab and GitHub: CI status, open MRs/PRs, failed job triage, and local checkout / worktree state. Uses `gh`, `glab`, and `git` on your machine.

  ## Context Branch
  This branch contains the project understanding and structural analysis for the `gitboard` repository. It serves as the foundation for future context updates.

  ## Install (binary)

  1. Install and log in to the forge CLIs:

  ```bash
  brew install glab
  glab auth login
  gh auth login
  ```

  2. Download the `gitboard` archive for your OS from the [latest release](https://github.com/behaviorengineering/gitboard/releases/latest), unpack it, and put `gitboard` on your `PATH`.

  3. Run:

  ```bash
  gitboard init
  # edit ~/.config/gitboard/config.yaml llm section if needed
  gitboard sync
  gitboard serve
  ```

  Open [http://127.0.0.1:1325/](http://127.0.0.1:1325/).

  ## From source

  ```bash
  brew install glab process-compose
  glab auth login
  gh auth login
  git clone https://github.com/behaviorengineering/gitboard.git
  cd gitboard
  make init
  # edit ~/.config/gitboard/config.yaml llm section if needed
  make sync
  make serve
  ```

  ## Config

  User config lives at `~/.config/gitboard/config.yaml` (or `$XDG_CONFIG_HOME/gitboard/config.yaml`).

  | Section | Purpose |
  |---------|---------|
  | `llm` | Optional AI triage (`base_url`, `model`, `api_key`) |
  | `ui` | Dashboard settings |
  | `local` | Disk roots to scan for checkouts / worktrees (`roots`) |
  | `sync` | GitHub orgs and GitLab groups to discover |
  | `projects` | Curated tracked repos |

  ## Commands

  | Command | Purpose |
  |---------|---------|
  | `gitboard init` | Create config if missing |
  | `gitboard sync` | Discover repos and select tracked projects |
  | `gitboard serve` | Direct dashboard (no process-compose) |
  | `make serve` | process-compose TUI on `:1325` |