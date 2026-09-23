# Gitboard ai-copilot skills

Canonical operator skills for this module. Host IDEs discover them via
[BOOTSTRAP.md](../BOOTSTRAP.md) symlinks (or junctions).

| Skill | Load when |
|-------|-----------|
| [gitboard-setup](gitboard-setup/SKILL.md) | First-time clone, config, serve, sync, `make ci`, conformity |
| [gitboard-docs](gitboard-docs/SKILL.md) | Adding or changing README, rules, skills, or operational docs |

## Related project guidance (not duplicated here)

| Path | Role |
|------|------|
| [`.cursor/rules/architecture.mdc`](../../.cursor/rules/architecture.mdc) | Packages, HTTP API, board behavior |
| [`.cursor/rules/always-rules-docs-as-code.mdc`](../../.cursor/rules/always-rules-docs-as-code.mdc) | Docs-as-code routing |
| [`.cursor/rules/always-rules-1-setup.mdc`](../../.cursor/rules/always-rules-1-setup.mdc) | Go module / workspace invariants |
| [`.cursor/skills/golang-quality/SKILL.md`](../../.cursor/skills/golang-quality/SKILL.md) | Go generation quality gates |
| [`.cursor/skills/live-conformity/SKILL.md`](../../.cursor/skills/live-conformity/SKILL.md) | Opt-in live forge tests |
| [`.cursor/skills/operator-config/SKILL.md`](../../.cursor/skills/operator-config/SKILL.md) | XDG config and secret patterns |
| [`.cursor/skills/release-gitboard/SKILL.md`](../../.cursor/skills/release-gitboard/SKILL.md) | Tag and GoReleaser release |

MUST NOT re-copy those bodies into `ai-copilots/`. Link and load them when needed.
