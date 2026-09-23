# BOOTSTRAP — Gitboard ai-copilots

**Audience:** Any AI agent (Cursor, GitHub Copilot, Claude Code, Codex) in a
workspace that depends on or checks out this module.

**Goal:** Wire host IDE discovery to canonical content under `ai-copilots/`.
Optionally refresh content. **MUST NOT** copy skill bodies unless symlinks or
junctions fail and the user approves copy fallback.

**Module path:** `github.com/behaviorengineering/gitboard`

---

## When to run

| Mode | Phases |
|------|--------|
| **Wire only** | 0 → 2 → 3 → 4 |
| **Refresh content + wire** | 0 → 1 → 2 → 3 → 4 |

---

## Phase 0 — Resolve module root

From a Go module that requires `github.com/behaviorengineering/gitboard` (or this checkout):

```bash
MOD="$(go list -m -f '{{.Dir}}' github.com/behaviorengineering/gitboard)"
test -d "$MOD/ai-copilots" || { echo "missing ai-copilots under $MOD"; exit 1; }
echo "Module Dir: $MOD"
```

If `go list` is unavailable, use a known nested checkout path only when it
clearly contains `ai-copilots/`. MUST NOT invent a path.

Re-run wire after module version bumps (cache Dir can change).

---

## Phase 1 — Refresh content (optional)

Edit only files under `$MOD/ai-copilots/`. Load author-ai-copilots + agent-smith
when authoring skills.

Target tree:

```text
ai-copilots/
  README.md
  BOOTSTRAP.md
  skills/gitboard-setup/SKILL.md
  skills/gitboard-docs/SKILL.md
```

MUST NOT put Gitboard-only operator content into `.cursor/packs/shared`
(cursor-packs). That pack is for portable shared practice across consumers.

---

## Phase 2 — Ask IDE and OS if unknown

1. IDE: Cursor, GitHub Copilot, Claude Code, Codex
2. OS: macOS/Linux symlink vs Windows junction/copy
3. Workspace: library alone vs nested under a parent monorepo vs dependency-only

---

## Phase 3 — Wire discovery

Canonical sources:

| Artifact | Path under `$MOD` |
|----------|-------------------|
| Skill tree | `ai-copilots/skills/gitboard-setup/` |
| Skill tree | `ai-copilots/skills/gitboard-docs/` |

Discovery paths:

| IDE | Agents | Skills |
|-----|--------|--------|
| Cursor | `.cursor/agents/*.md` | `.cursor/skills/**/SKILL.md` |
| GitHub Copilot | `.github/agents/*.agent.md` | `.github/skills/**/SKILL.md` |
| Claude Code | `.claude/agents/*.md` | `.claude/skills/**/SKILL.md` |
| Codex | `.codex/agents/*.md` | `.codex/skills/**/SKILL.md` |

**Cursor example (macOS/Linux)** from the **host workspace root** (this repo):

```bash
MOD="$(go list -m -f '{{.Dir}}' github.com/behaviorengineering/gitboard)"
mkdir -p .cursor/skills
ln -snf "$MOD/ai-copilots/skills/gitboard-setup" .cursor/skills/gitboard-setup
ln -snf "$MOD/ai-copilots/skills/gitboard-docs" .cursor/skills/gitboard-docs
```

When the workspace root is this checkout, relative links are fine:

```bash
ln -snf ../../ai-copilots/skills/gitboard-setup .cursor/skills/gitboard-setup
ln -snf ../../ai-copilots/skills/gitboard-docs .cursor/skills/gitboard-docs
```

**Windows:** prefer junction or developer-mode symlink; copy fallback only with
user approval.

**Idempotency:** skip if the link already resolves to the canonical path; ask
before overwriting stale copies.

**Parent monorepo:** some IDEs discover skills only at workspace-root
`.cursor/skills/`. Link there when the user wants root discoverability.
MUST NOT edit parent skill indexes unless the user asks.

---

## Phase 4 — Verify

```bash
MOD="$(go list -m -f '{{.Dir}}' github.com/behaviorengineering/gitboard)"
ls -la .cursor/skills/gitboard-setup .cursor/skills/gitboard-docs
test -f .cursor/skills/gitboard-setup/SKILL.md
test -f .cursor/skills/gitboard-docs/SKILL.md
test -f "$MOD/ai-copilots/BOOTSTRAP.md"
```

Ask the user before committing host wiring (`.cursor/`, `.github/`, etc.).
