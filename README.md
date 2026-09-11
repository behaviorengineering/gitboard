# Gitboard

<!-- majordomo-reading-nav:start -->
**Reading path:** [Next: mission.md](mission.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

<!-- majordomo-reading-toc:start -->
## Reading order

1. [README.md](README.md) (this file)
2. [mission.md](mission.md)
3. [architecture.md](architecture.md)
4. [conventions.md](conventions.md)
5. [weaknesses.md](weaknesses.md)
6. [chronology.md](chronology.md)
7. [evidence/typology/README.md](evidence/typology/README.md) - Typology seed briefing

<!-- majordomo-reading-toc:end -->

![Gitboard dashboard](docs/banner.webp)

Local code-change board for GitLab and GitHub: CI status, open MRs/PRs, failed job triage, and local checkout / worktree state. Uses `gh`, `glab`, and `git` on your machine.

## Context Branch

This is a context branch containing the project understanding for the `gitboard` repository. It is seeded from the `discover` evidence mode.

## Story vs Typology evidence

- Root markdown (`mission.md`, `architecture.md`, …) is the **teaching story**.
- `evidence/typology/` holds Typology **seed evidence** for this digest proposal.

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
