---
name: review-go-architecture
description: Verifies gitboard Go layout and dependency direction. Use after adding packages, moving code between internal/, or when the user asks for architecture review.
---

# Review Go Architecture (gitboard)

**Module:** `github.com/behaviorengineering/gitboard`. **Binary:** `bin/gitboard` (`cmd/gitboard`).

Invariants: `.cursor/rules/architecture.mdc`. This skill is the **procedure**.

---

## 1. Scope

Ask for target path(s). Default: `internal/` and `cmd/`.

---

## 2. Expected layout

```
cmd/gitboard/main.go
pkg/cliexec/
pkg/board/
pkg/remotegit/
pkg/localgit/
pkg/dashboard/
internal/config/
internal/syncproj/
internal/server/          # HTTP API + static UI
internal/triage/
internal/llm/
internal/pruneagent/      # strop agentsession (published module)
internal/observability/
internal/conformity/
internal/benchnet/
```

---

## 3. Dependency direction

| Layer | May import | Must not import |
|-------|------------|-----------------|
| `cmd/gitboard` | `internal/*`, `pkg/*` | — |
| `pkg/*` (portable domain) | other `pkg/*`; same-module `internal/config` when needed | `cmd` |
| `internal/*` (app) | `pkg/*`, other `internal/*` | reverse cycles into `pkg` that pull app UI |
| `server` | domain packages needed for handlers | reverse-import from domain into server cycles |

**Forbidden**

- Domain packages importing `cmd`
- Circular imports between packages- Growing business logic inside fat `main` switch arms (extract to domain)
- App features implemented by editing a vendored/submodule copy of strop

---

## 4. Checks

- [ ] New code sits in the right package
- [ ] No forbidden imports
- [ ] Shared process exec centralized
- [ ] HTTP routes registered in `internal/server` and match `architecture.mdc`
- [ ] Static UI stays under `internal/server/static` (or documented move)
- [ ] `make test` / `make vet` still pass

---

## 5. Report

List layout violations and suggested package moves. If user said **and fix**, apply safe moves and update imports.
