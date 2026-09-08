# Architecture

> **Teaching story.** Living project architecture for humans and review grounding.

Gitboard is organized into distinct bounded contexts (slices) that separate remote forge interaction, local filesystem management, and the user interface.

## Bounded Contexts

### Gitboard Core
Manages the end-to-end lifecycle of repository discovery, synchronization, and triage. It orchestrates the flow from remote discovery to actionable triage data.
- **Key Components:** `syncproj` (discovery), `triage` (processing), `cliexec` (execution utility).

### Dashboard
Provides the visual interface and data model for monitoring project status. It serves as the primary observation surface for the user.
- **Key Components:** `dashboard-ui` (web interface), `server` (API/HTTP surface), `board-dto` (data transfer objects).

### LocalGit
Manages local filesystem Git state and worktree inspections. It provides the bridge between the user's local disk and the project's data model.

### RemoteGit
Interfaces with GitHub and GitLab APIs to fetch remote repository metadata. This slice encapsulates the complexity of forge-specific API interactions.

### PruneAgent
An investigative service that analyzes local branches to identify likely removable or stale worktrees, supporting local environment hygiene.

## Technical Libraries
The following packages provide shared technical capabilities without domain-specific logic:
- **Config:** Centralized application configuration management.
- **LLM:** Integration capabilities for AI-assisted triage.
