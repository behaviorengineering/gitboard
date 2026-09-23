# Agents

This module is the Gitboard CLI and local dashboard. Humans read [README.md](README.md).

**Load these skills before you operate or extend this module:**

1. [ai-copilots/skills/README.md](ai-copilots/skills/README.md) (index)
2. [ai-copilots/skills/gitboard-setup/SKILL.md](ai-copilots/skills/gitboard-setup/SKILL.md) (setup, config, quality gates)
3. [ai-copilots/skills/gitboard-docs/SKILL.md](ai-copilots/skills/gitboard-docs/SKILL.md) (where docs and guidance belong)

**Also load when the task touches them:**

- [`.cursor/rules/architecture.mdc`](.cursor/rules/architecture.mdc) for packages, API, and board behavior
- [`.cursor/rules/always-rules-1-setup.mdc`](.cursor/rules/always-rules-1-setup.mdc) for Go module and workspace constraints
- [`.cursor/skills/golang-quality/SKILL.md`](.cursor/skills/golang-quality/SKILL.md) when writing or fixing Go
- Project review skills under [`.cursor/skills/`](.cursor/skills/) (`review-go-*`, `live-conformity`, `release-gitboard`) when those workflows apply

## Wire host discovery

Skills ship under `ai-copilots/`. They are not bound to one agent product.
Execute [ai-copilots/BOOTSTRAP.md](ai-copilots/BOOTSTRAP.md) in **wire mode**
to symlink into `.cursor/`, `.github/`, `.claude/`, or `.codex/`.

Resolve the module root when this library is only a Go dependency:

`go list -m -f '{{.Dir}}' github.com/behaviorengineering/gitboard`

MUST keep host links pointing at this module's `ai-copilots/` tree.
MUST NOT copy skill bodies into the host unless links fail and the user approves.
MUST NOT edit `.cursor/packs/shared` content as if it were Gitboard-owned.
MUST NOT overwrite unrelated working-tree changes when adding guidance files.
