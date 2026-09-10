# Chronology

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: weaknesses.md](weaknesses.md) · [Next: README.md](evidence/typology/README.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

  * **Seed**: Project understanding and structural typology established for `gitboard`.

### 2026-09-10 - majordomo - 3108088f0272

- **Did:** Add branch sync investigation modal for divergent locals. (#31)
- **Because:** Show ahead/behind commit evidence and allow fast-forward only when the checkout is clean and behind-only; hide inspect when there is nothing to investigate.
- **In order to:** advance context cursor on default first-parent tape
- **Evidence:** commit 3108088f02729934ac36589663cdcdcd002e7a0f; files: .cursor/rules/architecture.mdc, internal/dashboard/local.go, internal/dashboard/sync.go, internal/dashboard/sync_test.go, internal/localgit/prune.go, internal/localgit/prune_test.go, internal/localgit/sync.go, internal/localgit/sync_test.go, internal/server/server.go, internal/server/server_test.go, internal/server/static/app.css, internal/server/static/app.js, internal/server/static/index.html

### 2026-09-10 - majordomo - 1785e5ffb6b7

- **Did:** Show a blocked fast-forward control in dirty sync modals. (#32)
- **Because:** Keep pull available in the investigate dialog when behind-only, but disable it with a clear reason until the working tree is clean.
- **In order to:** advance context cursor on default first-parent tape
- **Evidence:** commit 1785e5ffb6b77b718872779f9763a4bf2085274a; files: internal/server/static/app.css, internal/server/static/app.js

### 2026-09-10 - majordomo - 0323fc3060cb

- **Did:** Update nested submodules after ff-only pulls. (#33)
- **Because:** Sync stale submodule checkouts before refusing a dirty tree and again after a successful fast-forward so pins like cursor-packs match the parent tip.
- **In order to:** advance context cursor on default first-parent tape
- **Evidence:** commit 0323fc3060cb23579ce4bd805b4f2109cf50494e; files: .cursor/rules/architecture.mdc, internal/localgit/pull.go, internal/localgit/pull_test.go
