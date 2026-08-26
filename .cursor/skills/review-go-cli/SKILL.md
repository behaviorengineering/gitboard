---
name: review-go-cli
description: Reviews gitboard CLI dispatch and command wiring in cmd/gitboard. Use when adding subcommands, changing flags, or after editing cmd/gitboard.
---

# Review Go CLI (gitboard)

**Spec sources:** `cmd/gitboard/main.go` (`printUsage`), `Makefile` (`bin/gitboard`).

---

## 1. Scope

Default: `cmd/gitboard/`. Narrow if the user names a subcommand only (`init`, `sync`, `serve`).

---

## 2. Entry pattern

- [ ] `main` dispatches known commands; unknown non-flag args → usage + exit 2.
- [ ] Backward-compatible: leading `-` args may still enter `serve`.
- [ ] No heavy business logic in `main` beyond flags + wiring.
- [ ] Domain work lives in `internal/*`.

---

## 3. Subcommands

### init

- [ ] Creates config if missing; never overwrites without explicit design change.
- [ ] Honors `-config` / default XDG path.

### sync

- [ ] Flags documented in usage / README commands table stay in sync.
- [ ] Interactive vs `--add` / `--remove` / `--dry-run` paths fail clearly.
- [ ] Writes `projects` via `internal/syncproj` / config helpers.

### serve

- [ ] Loads config; wires forge, localgit, dashboard, server, optional LLM/prune.
- [ ] Default addr is loopback.
- [ ] process-compose path remains `make serve` (scripts); CLI serve stays one-off.

---

## 4. Output and exits

- [ ] Errors to stderr / `log`; success paths quiet or intentional.
- [ ] Help via `-h` / `--help` / `help`.
- [ ] `printUsage` updated when commands or flags change.

---

## 5. Report

List CLI wiring issues. If user said **and fix**, apply them and keep usage strings aligned.
