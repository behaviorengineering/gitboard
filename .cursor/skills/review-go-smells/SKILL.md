---
name: review-go-smells
description: Identifies Go code smells, duplication, and maintainability issues in gitboard. Use when refactoring internal/ or cmd/, cleaning up technical debt, or when the user asks for smell review.
---

# Review Go Code Smells (gitboard)

**Goal:** Find maintainability issues without duplicating bug-hunt work. Pair with `review-go-code` for correctness.

---

## 1. Scope and tooling

Ask for target path(s). Default: `internal/`.

```bash
make test
make vet
gofmt -d ./cmd ./internal
```

Address tooling output first, then manual passes below.

---

## 2. Analysis steps

### 2.1 Duplication

- Repeated flag parsing, JSON encode/decode, or error-print patterns.
- Copy-pasted GitHub vs GitLab parsing that should share a helper in `forge`.
- Magic strings for API paths or config keys: prefer constants in one place.

### 2.2 Complexity

- Functions over ~80 lines: split by responsibility.
- Nesting deeper than 3 levels: early return or extract helper.
- Long parameter lists: use options struct.

### 2.3 Naming and packages

- Vague names (`data`, `utils`, `helper`, `manager`).
- Packages that mix unrelated concerns (for example HTTP + forge parsing in one file without need).
- Exported APIs that are only used inside the package (see `review-member-visibility`).

### 2.4 Comments and dead code

- Comments that restate the obvious or contradict code.
- Commented-out blocks; unused exports; unreachable branches.

### 2.5 Testing gaps that signal smell

- Untestable `exec.Command` buried in domain (should inject `cliexec`).
- Handlers that cannot be unit-tested without starting a full server when a smaller extract would do.

---

## 3. Report

Group by smell type. Prefer concrete refactors. If user said **and fix**, apply clear wins; ask before large moves.
