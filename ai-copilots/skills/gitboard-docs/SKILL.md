---
name: gitboard-docs
description: >-
  Route Gitboard documentation and AI guidance to the correct home: README
  pitch, .cursor rules for invariants, ai-copilots skills for operator
  procedures, and Makefile/tests for enforcement. Use when documenting setup,
  architecture, commands, or when asked to add CONTRIBUTING-style guidance.
---

# Gitboard docs

**Moral:** Operational knowledge becomes skills, rules, or executable checks.
README stays a short human pitch. Agents read `AGENTS.md` and `ai-copilots/`.

## When to load

- Adding or changing README, architecture notes, setup docs, or contributor guidance
- Choosing between README, `.cursor/rules/`, `ai-copilots/skills/`, and tests/Make
- Fixing stale command or API documentation after a behavior change
- Reviewing whether a new markdown file should exist at all

## Placement map

| Knowledge type | Where it lives |
|----------------|----------------|
| Human install + config + commands pitch | [README.md](../../../README.md) |
| Agent entry + load order | [AGENTS.md](../../../AGENTS.md) |
| Portable AI operator procedures | `ai-copilots/skills/*/SKILL.md` |
| Durable project invariants (packages, API, board) | `.cursor/rules/*.mdc` (especially `architecture.mdc`) |
| Shared cross-product practice | `.cursor/packs/shared` via cursor-packs workflow only |
| Repeatable checks | `Makefile`, Go/JS tests, CI workflows |

## Core constraints

**CONSTRAINT:** When documenting how an assistant should operate Gitboard, MUST
put the procedure in `ai-copilots/skills/` (and link from `AGENTS.md` / skill
index). MUST NOT grow a long standalone how-to under `docs/` that duplicates a
skill.

- Enforcement: new operational prose has a skill path or extends an existing skill
- Violation: STOP, move content into a skill, leave README as a short pointer

CORRECT:
```text
Setup steps → ai-copilots/skills/gitboard-setup/SKILL.md
AGENTS.md links to that skill
README keeps brew + make serve pitch
```

PROHIBITED:
```text
docs/contributor-setup.md with the full make/ci/serve runbook
# and no skill
```

**CONSTRAINT:** Durable layout, HTTP API, and board behavior MUST live in
[`.cursor/rules/architecture.mdc`](../../../.cursor/rules/architecture.mdc)
(or a focused project rule). MUST NOT dump architecture tables into README.

- Enforcement: README stays install/config/commands; architecture edits go to rules
- Violation: STOP, move tables to `architecture.mdc`

**CONSTRAINT:** Before changing package boundaries or public HTTP routes in
docs, MUST read `architecture.mdc` and align the rule with the code (or update
both together).

- Enforcement: docs claim matches `cmd/`, `internal/server`, and package layout
- Violation: STOP, fix the mismatch

**CONSTRAINT:** MUST NOT invent Make targets, CLI flags, env vars, or API routes
that are absent from the repository. When a doc change exposes a mismatch,
MUST update the doc to match code, or implement the missing surface with an
explicit user request.

- Enforcement: `make help`, `./bin/gitboard -h`, and route list in architecture
- Violation: STOP, correct the doc or ask before inventing behavior

**CONSTRAINT:** MUST NOT add Gitboard-only operator skills or product brand
paths into cursor-packs. Pack membership gate: shared practice only. Product
operator content stays in this module's `ai-copilots/` or host `.cursor/` overlay.

- Enforcement: `.cursor/rules/cursor-packs.mdc` three questions
- Violation: STOP, keep content under `ai-copilots/` or host overlay

**CONSTRAINT:** English only in code, identifiers, comments, paths, and docs.
MUST NOT use the em dash character (U+2014).

- Enforcement: scan new prose
- Violation: STOP, rewrite

## Steps

1. **Classify the request** — pitch vs invariant vs operator procedure vs enforceable check.
2. **Pick one home** from the placement map. Prefer extending an existing skill/rule over a new file.
3. **Cross-link** — `AGENTS.md` and `ai-copilots/skills/README.md` for skill load order; README for human commands only.
4. **Verify commands** — run or inspect `make help` / CLI usage / architecture API table before documenting them.
5. **Wire discovery** — if a new `ai-copilots` skill is added, update BOOTSTRAP examples and host symlinks per [BOOTSTRAP.md](../../BOOTSTRAP.md).

## Related

- Docs-as-code rule: `.cursor/rules/always-rules-docs-as-code.mdc`
- Setup skill: [gitboard-setup](../gitboard-setup/SKILL.md)
- Skill authoring standards: `.cursor/skills/agent-smith/SKILL.md`
- Pack edits: `.cursor/skills/edit-cursor-packs/SKILL.md`

## Pre-completion checklist

- [ ] **Correct home:** Content lives in README / rule / ai-copilot skill / Make-test as mapped
      Method: re-read placement map against the diff
      Pass: one clear home; no duplicate long guides
      Fail: STOP, relocate
- [ ] **No invented surface:** Commands and routes exist in tree
      Method: `make help` / CLI help / architecture.mdc
      Pass: every cited command/route is real
      Fail: STOP, fix doc or implement with explicit request
- [ ] **Architecture aligned:** Package/API claims match `architecture.mdc` and code
      Method: compare rule + code
      Pass: consistent
      Fail: STOP, update rule and/or code
- [ ] **No pack spill:** New Gitboard-only operator content is not in cursor-packs
      Method: path ownership check
      Pass: under `ai-copilots/` or host overlay
      Fail: STOP, move out of pack
- [ ] **Entry links:** New skills appear in `AGENTS.md` and `ai-copilots/skills/README.md`
      Method: open both files
      Pass: links resolve
      Fail: STOP, add index entries
