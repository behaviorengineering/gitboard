---
name: live-conformity
description: >-
  Run opt-in live forge conformity for gitboard: env vars, CLI credentials,
  make conformity-live, skip behavior, and how to read failures. Use when
  operating live network contract tests, GitHub Action live-conformity, or
  debugging GITBOARD_CONFORMITY_LIVE.
---

# Live conformity (gitboard)

**Moral:** Fake conformity is the default quality gate. Live tests hit real forges only when explicitly enabled, and they skip hosts that lack credentials or project paths.

## When to load

- Running or debugging `make conformity-live`
- Wiring or reading `.github/workflows/live-conformity.yml`
- Interpreting skipped hosts vs real Collect failures

## Enablement

**CONSTRAINT:** Live tests MUST run only when `GITBOARD_CONFORMITY_LIVE=1`. Ordinary `go test ./...`, `make conformity`, and PR CI MUST NOT set that variable.

- Enforcement: `TestLiveAdapterContract` skips unless the env equals `1`
- Violation: STOP, unset the env for default quality gates

CORRECT:
```bash
make conformity          # fake exec matrix
make conformity-live     # sets GITBOARD_CONFORMITY_LIVE=1
```

PROHIBITED:
```bash
# PR CI sets GITBOARD_CONFORMITY_LIVE=1 unconditionally
```

## Project path env (non-secret)

| Host | Env |
|------|-----|
| GitHub | `GITBOARD_LIVE_GITHUB_PATH` (`owner/repo`) |
| GitLab | `GITBOARD_LIVE_GITLAB_PATH` (`group/project`) |
| Azure DevOps | `GITBOARD_LIVE_AZURE_PATH` (`org/project/repo`) |
| Bitbucket | `GITBOARD_LIVE_BITBUCKET_PATH` (`workspace/repo`) |

Unset path → that host subtest skips.

## Credentials

| Host | How |
|------|-----|
| GitHub | `gh` installed and logged in (`gh auth login`, or `GH_TOKEN` / Action secret `GITBOARD_LIVE_GH_TOKEN`) |
| GitLab | `glab` installed and logged in (or Action secret `GITBOARD_LIVE_GITLAB_TOKEN`) |
| Azure DevOps | `az` installed and authenticated; PAT via `AZURE_DEVOPS_EXT_PAT` / secret `GITBOARD_LIVE_AZURE_PAT` |
| Bitbucket | `BITBUCKET_TOKEN` (Action secret `GITBOARD_LIVE_BITBUCKET_TOKEN`); no CLI login step |

**CONSTRAINT:** Operators MUST NOT print tokens, PATs, or full auth CLI stderr that may contain secrets. Live logs MUST report provider, project path, call count, and latency only.

- Enforcement: `live_test.go` logs `provider`, `project`, `calls`, `latency_ms`; sanitize auth detail
- Violation: STOP, remove credential echoes

## Commands

```bash
# Default (no network to forges)
make conformity

# Live opt-in (skips hosts without path/creds)
export GITBOARD_LIVE_GITHUB_PATH=owner/repo
# optional: GITBOARD_LIVE_GITLAB_PATH, GITBOARD_LIVE_AZURE_PATH, GITBOARD_LIVE_BITBUCKET_PATH
# optional: BITBUCKET_TOKEN, az/gh/glab already logged in
make conformity-live
```

GitHub Action: workflow_dispatch or Monday schedule. Supply paths via workflow inputs or repository variables. Inject credentials only from secrets.

## Skip vs fail

| Symptom | Meaning |
|---------|---------|
| `GITBOARD_CONFORMITY_LIVE` skip | Live mode off (expected in PR CI) |
| `GITBOARD_LIVE_*_PATH unset` | Host not configured for this run |
| `credentials unavailable` | CLI missing or not authenticated; not a product regression |
| `collect:` / `project error:` / `RemoteNamesOK=false` | Real adapter or auth failure against the configured project |

## Checklist

- [ ] Live env is opt-in only
- [ ] Paths are non-secret env/vars; tokens are secrets
- [ ] Logs never print credentials
- [ ] Fake `make conformity` remains the default gate
