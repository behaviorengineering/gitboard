# Gitboard

![Gitboard dashboard](docs/banner.webp)

Local code-change board for GitLab and GitHub: CI status, open MRs/PRs, failed job triage, and local checkout / worktree state. Uses `gh`, `glab`, and `git` on your machine (not deployments or runtime ops).

This context branch is seeded from the `604a237` discovery pass.

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

Open [http://127.0.0.1:1325/](http://127.0.0.1:1325/). Each GitHub Release includes notes for what changed since the previous tag.

`process-compose` is only needed for the from-source `make serve` TUI below.

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

Open [http://127.0.0.1:1325/](http://127.0.0.1:1325/). Stop with `make serve-down`.

`make serve` runs a **process-compose** TUI. The CLI `gitboard serve` still works for one-off runs without the TUI.

## Config

User config lives at `~/.config/gitboard/config.yaml` (or `$XDG_CONFIG_HOME/gitboard/config.yaml`).

| Section | Purpose |
|---------|---------|
| `llm` | Optional AI triage (`base_url`, `model`, `api_key`) |
| `ui` | Dashboard settings (see below) |
| `local` | Disk roots to scan for checkouts / worktrees (`roots`) |
| `sync` | GitHub orgs and GitLab groups to discover |
| `projects` | Curated tracked repos (written by `sync`; optional `local_path`) |

`ui` fields:

| Key | Purpose |
|-----|---------|
| `poll_seconds` | Browser auto-refresh interval (default 30; `0` disables) |
| `hide_branches` | Optional `path.Match` patterns; matching branch names are omitted from the Branches column only (prune / forge data unchanged). Empty or omitted shows all. |

```yaml
ui:
  poll_seconds: 30
  hide_branches:
    - majordomo-context/*
    - dependabot/*
```

Override path with `GITBOARD_CONFIG`. LLM fields also accept env overrides: `GITBOARD_LLM_BASE_URL`, `GITBOARD_LLM_MODEL`, `GITBOARD_LLM_API_KEY`, `POLYPUS_BASE_URL`, `OPENAI_API_KEY`. Poll interval: `GITBOARD_POLL_SECONDS`.

## Commands

| Command | Purpose |
|---------|---------|
| `gitboard init` / `make init` | Create config if missing (never overwrite) |
| `gitboard sync` / `make sync` | Discover repos and select tracked projects |
| `make serve` | process-compose TUI on `:1325`; rebuilds/restarts when `.go` or static UI files change |
| `make serve-down` | Stop this Gitboard process-compose project |
| `gitboard serve` | Direct dashboard (no process-compose) |
| `gitboard version` | Print the build version |

Sync helpers: `--add owner/repo --host github|gitlab`, `--remove id`, `--dry-run`, `--host github|gitlab`.

`make sync` uses an interactive TUI (Charm huh): arrow keys, space to toggle, `/` to filter, enter to confirm. Already-tracked repos start checked.

## Local checkouts

Set `local.roots` to directories that contain clones (including nested submodules and linked worktrees). Gitboard matches `origin` remotes to tracked `host`/`path`.

Per-project override:

```yaml
projects:
  - id: gitboard
    label: gitboard
    host: github
    path: behaviorengineering/gitboard
    local_path: ~/Xynova/ai/gitboard
```