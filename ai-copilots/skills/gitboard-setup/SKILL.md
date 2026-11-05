---
name: gitboard-setup
description: >-
  Set up and operate the Gitboard Go module: prerequisites, XDG config,
  go tool task ci, serve, sync, conformity, and secret-safe environment overrides.
  Use when cloning, bootstrapping, configuring, serving, syncing projects,
  or verifying local quality gates.
---

# Gitboard setup

**Moral:** Treat `go tool task ci` as the local quality gate. Rely on forge CLI sessions
for auth. Keep secrets out of YAML literals.

## When to load

- Fresh clone or contributor onboarding
- Config init, sync, serve, or environment override questions
- Running tests, lint, conformity, or diagnosing a failed task
- Before claiming a Go or dashboard change is ready

## Prerequisites

**CONSTRAINT:** Before `go tool task serve` or forge-backed sync, MUST ensure required tools exist.

| Need | Tools |
|------|-------|
| Always (from source) | Go (see `go.mod`), `git`, Node.js (dashboard JS tests) |
| Binary / forge UI | `gh`, `glab` logged in (`gh auth login`, `glab auth login`) |
| `go tool task serve` TUI | `process-compose` |
| Lint, format, task runner | none on PATH: golangci-lint, gofumpt and task are pinned in `go.mod` `tool` |
| Azure DevOps adapter | `az` authenticated |
| Bitbucket adapter | `BITBUCKET_TOKEN` or username/app-password env (no interactive CLI login) |

- Enforcement: `command -v` / `go version` / `node --version` before dependent targets
- Violation: STOP, install missing tools, then continue

CORRECT:
```bash
brew install gh glab process-compose
gh auth login
glab auth login
```

PROHIBITED:
```bash
# assume gh/glab sessions exist without checking
go tool task sync
```

## Core commands

The runner is `Taskfile.yml`, run through the pinned task binary (`go tool task`).
Pass extra CLI flags after `--`.

```bash
go tool task                 # list local-dev tasks
go tool task build
go tool task test            # JS + Go unit tests
go tool task test-js         # unlisted helper
go tool task vet
go tool task format          # gofumpt
go tool task lint            # whole module, including older debt
go tool task tidy
go tool task ci              # primary local gate
go tool task init            # create config if missing
go tool task sync            # interactive project selection
go tool task sync -- -host github -dry-run
go tool task serve           # process-compose on :1325
go tool task serve-down
go tool task conformity          # fake-exec network contracts
go tool task conformity-live     # opt-in live forge Collect
go tool task bench-net           # unlisted microbenchmarks
```

Direct CLI (after `go tool task build` or installed binary):

```bash
./bin/gitboard init
./bin/gitboard sync [-host …] [-add owner/repo -host …] [-remove id] [-dry-run]
./bin/gitboard serve [-addr 127.0.0.1:1325] [-allow-non-localhost]
./bin/gitboard version
```

**CONSTRAINT:** Agents MUST use `go tool task ci` (tidy check, gofumpt check, vet,
golangci-lint on lines changed since `origin/main`, JS tests, race tests, build)
before claiming a change is complete.

- Enforcement: run `go tool task ci` after substantive edits
- Violation: STOP, run the gate, fix failures

## Config

Default user path: `~/.config/gitboard/config.yaml` (or
`$XDG_CONFIG_HOME/gitboard/config.yaml`).

Binary resolution order (`internal/config.DefaultPath`):

1. `GITBOARD_CONFIG`
2. legacy `GITBOARD_PROJECTS`
3. existing user `config.yaml`
4. cwd `config.yaml` / `projects.yaml`
5. user config path (may not exist yet)

`gitboard init` and `go tool task serve` create the file if missing and never overwrite
an existing file. New files use restrictive permissions (`0600`).

**CONSTRAINT:** MUST NOT write literal API keys, tokens, or passwords into
`config.yaml`, examples, or tests. Use `${VAR}` placeholders and environment
overrides.

- Enforcement: scan new YAML/examples for secret-shaped literals
- Violation: STOP, replace with env placeholders

Supported overrides include:

```text
GITBOARD_CONFIG
GITBOARD_LLM_BASE_URL
GITBOARD_LLM_MODEL
GITBOARD_LLM_API_KEY
POLYPUS_BASE_URL
OPENAI_API_KEY
GITBOARD_POLL_SECONDS
GITBOARD_OPENINFERENCE_ENABLED
GITBOARD_OPENINFERENCE_ENDPOINT
```

**Note:** Tasks pass no config path, so `init`, `sync` and `serve` use the
binary's resolution order above. Override with `GITBOARD_CONFIG=…` or flags after
`--`, for example `go tool task init -- -config path`.

Human overview of config sections: [README.md](../../../README.md).
Deeper operator-config patterns: `.cursor/skills/operator-config/SKILL.md`.

## Serve and sync

**CONSTRAINT:** The dashboard MUST bind to loopback by default. MUST NOT widen
the listen address without an explicit user request and `-allow-non-localhost`.

- Enforcement: `gitboard serve` refuses non-loopback unless the flag is set
- Violation: STOP, keep `127.0.0.1` / localhost

```bash
go tool task serve       # builds, creates config if missing, process-compose TUI
go tool task serve-down
./bin/gitboard serve     # one-off, no process-compose
go tool task sync        # interactive Charm selector (GitHub + GitLab sources in UI)
```

Open [http://127.0.0.1:1325/](http://127.0.0.1:1325/) after serve starts.

## Conformity and releases

```bash
go tool task conformity
# Live only when intentionally enabled:
go tool task conformity-live
```

Live details: `.cursor/skills/live-conformity/SKILL.md`.

Tag-driven releases: `.cursor/skills/release-gitboard/SKILL.md`.

## Module and dependency constraints

**CONSTRAINT:** MUST NOT run `go mod init`, delete `go.mod` / `go.sum` /
`go.work`, re-add a long-lived `providers/strop` submodule, or permanently
`replace` strop unless the user explicitly asks.

- Enforcement: `.cursor/rules/always-rules-1-setup.mdc`
- Violation: STOP, restore module files, use `go get` + `go mod tidy`

After dependency edits: `go mod tidy`, then `go mod verify` when checking
integrity. Prefer published `github.com/behaviorengineering/strop@vX.Y.Z`.

## Related skills

| Task | Load |
|------|------|
| Go generation / quality | `.cursor/skills/golang-quality/SKILL.md` |
| CLI surface | `.cursor/skills/cli-command-surface/SKILL.md` |
| Architecture | `.cursor/rules/architecture.mdc` |
| Docs placement | [gitboard-docs](../gitboard-docs/SKILL.md) |

## Pre-completion checklist

- [ ] **Tools:** Required CLIs for the task are installed and authenticated
      Method: `command -v` / `gh auth status` / `glab auth status` as needed
      Pass: binaries present; auth OK for forge work
      Fail: STOP, install or log in
- [ ] **Config:** No literal secrets; path resolution understood
      Method: inspect config/examples; prefer env overrides
      Pass: placeholders or env only
      Fail: STOP, remove secrets
- [ ] **Gate:** `go tool task ci` run after substantive changes
      Method: execute the task
      Pass: exit 0
      Fail: STOP, fix failures
- [ ] **Serve bind:** still loopback unless user requested otherwise
      Method: check `-addr` / `-allow-non-localhost` usage
      Pass: localhost default preserved
      Fail: STOP, revert bind widening
