---
name: gitboard-setup
description: >-
  Set up and operate the Gitboard Go module: prerequisites, XDG config,
  make ci, serve, sync, conformity, and secret-safe environment overrides.
  Use when cloning, bootstrapping, configuring, serving, syncing projects,
  or verifying local quality gates.
---

# Gitboard setup

**Moral:** Treat `make ci` as the local quality gate. Rely on forge CLI sessions
for auth. Keep secrets out of YAML literals.

## When to load

- Fresh clone or contributor onboarding
- Config init, sync, serve, or environment override questions
- Running tests, lint, conformity, or diagnosing a failed `make` target
- Before claiming a Go or dashboard change is ready

## Prerequisites

**CONSTRAINT:** Before `make serve` or forge-backed sync, MUST ensure required tools exist.

| Need | Tools |
|------|-------|
| Always (from source) | Go (see `go.mod`), `git`, Node.js (dashboard JS tests) |
| Binary / forge UI | `gh`, `glab` logged in (`gh auth login`, `glab auth login`) |
| `make serve` TUI | `process-compose` |
| `make lint` | `golangci-lint` |
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
make sync
```

## Core commands

```bash
make help
make build
make test          # JS + Go unit tests
make test-js
make vet
make format
make lint          # needs golangci-lint
make tidy
make ci            # primary local gate
make init          # create config if missing
make sync          # interactive project selection
make serve         # process-compose on :1325
make serve-down
make conformity          # fake-exec network contracts
make conformity-live     # opt-in live forge Collect
make bench-net
```

Direct CLI (after `make build` or installed binary):

```bash
./bin/gitboard init
./bin/gitboard sync [-host …] [-add owner/repo -host …] [-remove id] [-dry-run]
./bin/gitboard serve [-addr 127.0.0.1:1325] [-allow-non-localhost]
./bin/gitboard version
```

**CONSTRAINT:** Agents MUST use `make ci` (or the equivalent tidy + gofmt check +
vet + race tests + build + JS tests) before claiming a change is complete.

- Enforcement: run `make ci` after substantive edits
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

`gitboard init` and `make serve` create the file if missing and never overwrite
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

**Note:** Makefile targets use `CONFIG ?= $(HOME)/.config/gitboard/config.yaml`.
That Make default does not automatically honor `XDG_CONFIG_HOME` or
`GITBOARD_CONFIG`. Prefer passing `CONFIG=…` to Make, or run `./bin/gitboard …`
with the binary's path resolution, when those overrides matter.

Human overview of config sections: [README.md](../../../README.md).
Deeper operator-config patterns: `.cursor/skills/operator-config/SKILL.md`.

## Serve and sync

**CONSTRAINT:** The dashboard MUST bind to loopback by default. MUST NOT widen
the listen address without an explicit user request and `-allow-non-localhost`.

- Enforcement: `gitboard serve` refuses non-loopback unless the flag is set
- Violation: STOP, keep `127.0.0.1` / localhost

```bash
make serve          # builds, creates config if missing, process-compose TUI
make serve-down
./bin/gitboard serve   # one-off, no process-compose
make sync              # interactive Charm selector (GitHub + GitLab sources in UI)
```

Open [http://127.0.0.1:1325/](http://127.0.0.1:1325/) after serve starts.

## Conformity and releases

```bash
make conformity
# Live only when intentionally enabled:
make conformity-live
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
- [ ] **Gate:** `make ci` (or equivalent) run after substantive changes
      Method: execute the target
      Pass: exit 0
      Fail: STOP, fix failures
- [ ] **Serve bind:** still loopback unless user requested otherwise
      Method: check `-addr` / `-allow-non-localhost` usage
      Pass: localhost default preserved
      Fail: STOP, revert bind widening
