---
name: review-go-code
description: Systematic Go bug review for gitboard. Traces execution flow, resources, logic, and forge/CLI safety. Use after implementing or refactoring Go in cmd/ or internal/, or when the user asks for a code review or bug hunt.
---

# Review Go Code (gitboard)

**Module:** `github.com/behaviorengineering/gitboard`  
**Goal:** Find bugs before merge. Focus on correctness, resource leaks, and CLI/forge failure handling.

Style: `.cursor/rules/rules-for-golang-coding.mdc`. Layout: `.cursor/skills/review-go-architecture/SKILL.md`.

---

## 1. Scope

Ask for target path(s). Default: changed files, or `internal/` if unclear.

```bash
make test
make vet
```

Fix or report failures before manual review.

---

## 2. Review steps

Use `checklist.md` for pattern-level checks.

### 2.1 Execution flow

- Trace entry: `cmd/gitboard/main.go` → `runInit` / `runSync` / `runServe` → `internal/*`.
- Exit codes: success 0; usage/unknown command 2 where applicable; other errors non-zero.
- Flag parsing stays in cmd; domain logic in `internal/*`.
- Follow each error return; no silent success on partial failure.

### 2.2 Data and process flow

- Forge/git process calls go through `internal/cliexec` or existing forge/localgit helpers.
- Config paths respect `GITBOARD_CONFIG` / XDG defaults from `internal/config`.
- Local paths: expand `~` consistently; do not assume cwd is a project root.

### 2.3 Resources

- Every `os.Open` / `os.Create` has `defer Close()` (or explicit close on all paths).
- `bufio.Scanner`: check `scanner.Err()` after the loop.
- HTTP handlers: context cancellation respected where long work runs.
- Temp files removed on error paths.

### 2.4 Logic and edge cases

- Empty project lists, missing config, missing `gh`/`glab`: clear errors, no panic.
- Slice/map access bounded; nil checks before dereference.
- Named return `(err error)` with `defer`: use `err =` not `err :=` when defer reads `err`.

### 2.5 Security and safety

- No tokens or API keys in source or fixtures.
- Server listen address stays localhost unless the user explicitly asked to widen it.
- Destructive local ops (`prune/safe`, delete checkout) re-check safety hints before acting.
- Pull is ff-only; never force-push or hard-reset from the dashboard.

---

## 3. Report

List findings by severity (blocker / should-fix / nit). Cite file:line. Say what to change. If the user said **and fix**, apply blocker and should-fix items.
