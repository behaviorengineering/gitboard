---
name: release-gitboard
description: >-
  Cut a gitboard GitHub Release via semver tag and GoReleaser. Use when tagging
  a release, publishing binaries, or checking release notes after a version bump.
---

# Release gitboard

**Spec sources:** `.goreleaser.yaml`, `.github/workflows/release.yml`, `README.md` (Install binary).

## When to invoke

- User asks to release, tag, or publish gitboard binaries.
- After merge to `main` when cutting `vMAJOR.MINOR.PATCH`.

## Must

- Release only from green `main` (CI quality job passed).
- Use annotated semver tags: `v0.1.0`, `v0.2.0`, …
- Push the tag to `origin` so `release.yml` runs GoReleaser.
- Confirm the GitHub Release has binaries, `checksums.txt`, and changelog groups (Features / Bug fixes / Docs / Others).

## Must not

- Do not force-push or move an existing release tag.
- Do not create tags from dirty or unmerged feature branches.
- Do not skip CI by tagging a commit that never passed `make ci` / Actions.
- Do not add Homebrew or Docker in this skill unless the user asks.

## Commands

```bash
# 1. On main, clean tree, CI green
git checkout main
git pull --ff-only
git status

# 2. Annotated tag (example)
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0

# 3. Watch the Release workflow
gh run watch
gh release view v0.1.0
```

Local config check (no publish):

```bash
goreleaser check
```

## Notes

- GoReleaser sets `main.version` via ldflags; `gitboard version` should match the tag.
- Changelog comes from GitHub compare since the previous tag (see `.goreleaser.yaml`).
- Contributors keep using `make serve`; end users download the Release binary.
