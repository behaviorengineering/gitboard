# Architecture

Gitboard is organized into functional slices that separate domain logic from delivery mechanisms (CLI and Web UI).

## Bounded Contexts

### Core Domains
- **Board**: Manages high-level project state and the visual representation of repository health.
- **Config**: Centralized management of repository settings and tool configurations.
- **LocalGit**: Handles local filesystem operations and worktree inspections.
- **Remotegit**: Manages interactions with remote hosting providers (GitHub/GitLab).
- **Triage**: Analyzes repository state to prioritize maintenance and cleanup.
- **LLM**: Provides generative AI capabilities for repository analysis and triage assistance.
- **PruneAgent**: Automates decision-making for cleaning up stale branches and resources.

### Delivery &amp; Interface
- **Gitboard (CLI)**: The primary orchestration layer for command-line tasks, including observability and server management.
- **Dashboard (UI)**: The web-based interface for monitoring project status and git activity.

## Component Mapping

| Slice | Primary Packages | Surface |
|-------|------------------|---------|
| `board` | `./internal/board` | - |
| `config` | `./internal/config` | - |
| `dashboard` | `./internal/dashboard` | `dashboard-ui` |
| `gitboard` | `./cmd/gitboard`, `./internal/observability`, `./internal/server`, `./internal/syncproj` | `gitboard-cli` |
| `localgit` | `./internal/localgit` | - |
| `llm` | `./internal/llm` | - |
| `pruneagent` | `./internal/pruneagent` | - |
| `remotegit` | `./internal/remotegit` | - |
| `triage` | `./internal/triage` | - |