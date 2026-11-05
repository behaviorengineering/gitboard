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

## Wire host discovery

Cursor loads fleet skills and rules from the cursor-packs plugin. This module
keeps `.cursor/packs/shared` for Lefthook and CI scripts only.

Skills ship under `ai-copilots/`. They are not bound to one agent product.
Execute [ai-copilots/BOOTSTRAP.md](ai-copilots/BOOTSTRAP.md) in **wire mode**
to symlink into `.cursor/`, `.github/`, `.claude/`, or `.codex/`.

Resolve the module root when this library is only a Go dependency:

`go list -m -f '{{.Dir}}' github.com/behaviorengineering/gitboard`

MUST keep host links pointing at this module's `ai-copilots/` tree.
MUST NOT copy skill bodies into the host unless links fail and the user approves.
MUST NOT edit `.cursor/packs/shared` or the cursor-packs plugin checkout as if they were Gitboard-owned.
MUST NOT dual-run pack-named `.cursor/` skill or rule symlinks alongside the plugin.
MUST NOT overwrite unrelated working-tree changes when adding guidance files.
