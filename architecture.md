<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: mission.md](mission.md) · [Next: conventions.md](conventions.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

markdown: |
# Architecture
> Teaching story. Living project architecture for humans and review grounding. Not Typology seed evidence (see `evidence/typology/architecture_brief.md`).
The system operates as a local orchestration layer for forge-based workflows. It integrates directly with existing CLI tools—`gh`, `glab`, and `git`—to aggregate state from GitLab and GitHub without requiring server-side deployments.
The architecture centers on a synchronization loop that pulls remote repository status (CI status, open MRs) and maps it against local worktree states. This creates a unified dashboard view that bridges the gap between remote forge activity and local developer context. The design prioritizes local-first execution, ensuring that the tool acts as a transparent observer and facilitator of the developer's existing toolchain.
