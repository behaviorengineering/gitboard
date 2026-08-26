---
name: review-member-visibility
description: Audits Go packages for visibility (export only what's essential). Use when the user asks to review visibility, audit exports, or reduce public API surface.
---

# Review Member Visibility

**Principle:** Export only what's essential (public API interfaces, constructors, shared domain/DTOs). Keep implementation structs and internal helpers private.

---

## 1. Scope

Default: `internal/`. Or the path the user gave (e.g. `internal/forge`, `internal/server`).

---

## 2. List exported symbols

Per package in scope:

```bash
rg -n "^type [A-Z]" <pkg_path> --glob '*.go'
rg -n "^func [A-Z]|^var [A-Z]|^const [A-Z]" <pkg_path> --glob '*.go'
```

Record: package, symbol, file:line, kind.

---

## 3. Decision tree (per symbol)

1. **Used outside this package?** (Search for `pkg.Symbol` outside the package.) **NO** → private. **YES** → continue.
2. **Public API?** (Interface callers use, constructor, shared DTO.) **NO** → private. **YES** → continue.
3. **Constructor (New*)?** → keep public.
4. **Interface for DI/testing or shared DTO?** → keep public.
5. **Otherwise** → private.

**Keep public:** Interfaces as param/return, DTOs used elsewhere, constructors.  
**Make private:** Impl structs (callers use interface + New*), internal helpers, internal interfaces, vars only for building exports.

---

## 4. Report

Table: symbol | keep/unexport | reason. If user said **and fix**, unexport safe symbols and update call sites inside the package.
