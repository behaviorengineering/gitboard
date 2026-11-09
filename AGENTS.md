# Agents

This module is the Gitboard CLI and local dashboard. Humans read [README.md](README.md).

**Load these skills before you operate or extend this module:**

1. [ai-copilots/skills/README.md](ai-copilots/skills/README.md) (index)
2. [ai-copilots/skills/gitboard-setup/SKILL.md](ai-copilots/skills/gitboard-setup/SKILL.md) (setup, config, quality gates)
3. [ai-copilots/skills/gitboard-docs/SKILL.md](ai-copilots/skills/gitboard-docs/SKILL.md) (where docs and guidance belong)

**Also load when the task touches them:**

- [`.cursor/rules/architecture.mdc`](.cursor/rules/architecture.mdc) for packages, API, and board behavior
- cursor-packs plugin skill `golang-quality` when writing or fixing Go
- cursor-packs plugin skills `review-code-staged`, `review-code-smells`, `review-member-visibility`, and `cli-command-surface` for reviews
- Project skills under [`.cursor/skills/`](.cursor/skills/) (`live-conformity`, `release-gitboard`) when those workflows apply

## Operate the CLI

- `gitboard` prints the agent operating guide (config discovery order, exit codes); `gitboard help` lists commands.
- Read-only: `gitboard version`, `gitboard sync --dry-run`, `gitboard init --dry-run`. Dry runs write nothing, not even a missing config file.
- Mutating: `gitboard init`, `gitboard sync --host <host> --add owner/repo`, `gitboard sync --remove <id>`. Interactive `gitboard sync` needs a terminal; without one it exits 2 and names the flags.
- Long-running: `gitboard serve` (127.0.0.1:1325) or `go tool task serve`. With `--allow-non-localhost`, mutating API calls need the token printed at startup in `X-Gitboard-Token`, and the board shows a fixed message instead of raw error, warning and tooling detail (the detail goes to the server log).

Exit codes (from the `pkg/gberr` code of the returned error):

| Code | Meaning |
|------|---------|
| 0 | success |
| 1 | internal error (also any error without a code) |
| 2 | invalid input or usage: bad flag or argument, unknown host or project, unreadable config, no terminal |
| 3 | conflict with the current state (dirty tree, diverged branch, confirmation required) |
| 4 | not found (for example no repositories for the sync sources) |
| 5 | a forge, git or the network is unavailable |
| 6 | cancelled or timed out |

## Wire host discovery

Cursor loads fleet skills and rules from the cursor-packs plugin. Lefthook
lives in this repo (`lefthook.yml` extends `lefthook/lefthook-go-standalone.yml`).

Skills ship under `ai-copilots/`. They are not bound to one agent product.
Execute [ai-copilots/BOOTSTRAP.md](ai-copilots/BOOTSTRAP.md) in **wire mode**
to symlink into `.cursor/`, `.github/`, `.claude/`, or `.codex/`.

Resolve the module root when this library is only a Go dependency:

`go list -m -f '{{.Dir}}' github.com/behaviorengineering/gitboard`

MUST keep host links pointing at this module's `ai-copilots/` tree.
MUST NOT copy skill bodies into the host unless links fail and the user approves.
MUST NOT edit the cursor-packs plugin checkout as if it were Gitboard-owned.
MUST NOT dual-run pack-named `.cursor/` skill or rule symlinks alongside the plugin.
MUST NOT overwrite unrelated working-tree changes when adding guidance files.
