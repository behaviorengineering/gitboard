# Gitboard

![Gitboard dashboard](docs/banner.webp)

Local code-change board for CI status, open PRs/MRs, failed-job triage, and local checkout / worktree state. Uses `gh`, `glab`, and `git` on your machine (not deployments or runtime ops).

Adapters support GitHub, GitLab, Azure DevOps, and Bitbucket. Manage Sources/Discover and interactive `gitboard sync` currently expose GitHub and GitLab only.

## Install (binary)

1. Install and log in to the forge CLIs:

```bash
brew install gh glab
glab auth login
gh auth login
```

2. Download the `gitboard` archive for your OS from the [latest release](https://github.com/behaviorengineering/gitboard/releases/latest), unpack it, and put `gitboard` on your `PATH`. (`go install github.com/behaviorengineering/gitboard/cmd/gitboard@vX.Y.Z` also works; `gitboard version` reports the module tag via BuildInfo.)

3. Run:

```bash
gitboard serve
```

The first run creates `~/.config/gitboard/config.yaml` if missing (never overwrites). Open [http://127.0.0.1:1325/](http://127.0.0.1:1325/), then use **Manage** (or `gitboard sync`) to add projects. Optional: edit the `llm` section for AI triage.

Each GitHub Release includes notes for what changed since the previous tag.

`process-compose` is only needed for the from-source `make serve` TUI below.

## From source

```bash
brew install gh glab process-compose
glab auth login
gh auth login
git clone https://github.com/behaviorengineering/gitboard.git
cd gitboard
make serve
```

Open [http://127.0.0.1:1325/](http://127.0.0.1:1325/). Stop with `make serve-down`. `make serve` and `gitboard serve` both create the config file if it is missing.

`make serve` runs a **process-compose** TUI. The CLI `gitboard serve` still works for one-off runs without the TUI.

## Config

User config lives at `~/.config/gitboard/config.yaml` (or `$XDG_CONFIG_HOME/gitboard/config.yaml`).

| Section | Purpose |
|---------|---------|
| `llm` | Optional AI triage (`base_url`, `model`, `api_key`) |
| `openinference` | Optional tracing and inference failure dumps |
| `ui` | Dashboard settings (see below) |
| `local` | Disk roots to scan for checkouts / worktrees (`roots`, `fetch_seconds`, `scan_seconds`) |
| `sync` | Discovery sources for GitHub, GitLab, Azure DevOps, and Bitbucket |
| `projects` | Curated tracked repos (written by `sync`; optional `local_path`) |
| `views` | Named board subsets of project ids (optional; omit for one default view over all projects) |

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

Edit projects, views, roots, or `ui.hide_branches` while `serve` is running. The server reloads the config when the file changes; no restart required.

Board pulls are fast-forward only. After a pull, Gitboard runs `git submodule update --init --recursive` on the checked-out worktree.

Override path with `GITBOARD_CONFIG`. LLM fields also accept env overrides: `GITBOARD_LLM_BASE_URL`, `GITBOARD_LLM_MODEL`, `GITBOARD_LLM_API_KEY`, `POLYPUS_BASE_URL`, `OPENAI_API_KEY`. Poll interval: `GITBOARD_POLL_SECONDS`. OpenInference: `GITBOARD_OPENINFERENCE_ENABLED`, `GITBOARD_OPENINFERENCE_ENDPOINT`.

## Commands

| Command | Purpose |
|---------|---------|
| `gitboard init` / `make init` | Create config if missing (never overwrite) |
| `gitboard sync` / `make sync` | Discover repos and select tracked projects |
| `make serve` | process-compose TUI on `:1325`; rebuilds/restarts when `.go` or static UI files change |
| `make serve-down` | Stop this Gitboard process-compose project |
| `gitboard serve` | Direct dashboard (no process-compose) |
| `gitboard version` | Print the build version |

Sync helpers: `-add owner/repo -host github|gitlab|azuredevops|bitbucket`, `-remove id`, `-dry-run`, `-host github|gitlab|azuredevops|bitbucket`.

`make sync` uses an interactive TUI (Charm huh): arrow keys, space to toggle, `/` to filter, enter to confirm. Already-tracked repos start checked. The interactive prompt currently lists GitHub and GitLab sources.

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
