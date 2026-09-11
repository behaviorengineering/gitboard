# Architecture

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: mission.md](mission.md) · [Next: conventions.md](conventions.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

> **Teaching story.** Living project architecture for humans and review grounding. Not Typology seed evidence (see `evidence/typology/architecture_brief.md`).

Gitboard is structured as a collection of bounded contexts that separate the concerns of data ingestion, local state inspection, and user presentation.

## Bounded Contexts

### Core Data &amp; Domain
- **board**: Defines the central JSON data shapes (DTOs) shared across the dashboard UI and forge adapters.
- **cliexec**: A low-level runner interface for executing external processes (e.g., `git`, `gh`, `glab`).

### Forge &amp; Remote Integration
- **remotegit**: Adapts external Git provider data into internal `board` DTOs.
- **syncproj**: Aggregates project information from external forges like GitHub and GitLab.

### Local State &amp; Intelligence
- **localgit**: Inspects local git worktrees and synchronizes branch states.
- **pruneagent**: Investigates and decides the lifecycle status of local-only or removable branches.
- **triage**: Provides an Analyzer to process job and log information via a Request/Response pattern, potentially utilizing LLMs.
- **llm**: Provides an OpenAI-compatible client to interface with external LLM services for intelligence tasks.

### Presentation &amp; Infrastructure
- **server**: Provides HTTP services and initializes request multiplexing for the web UI.
- **dashboard**: Collects and investigates synchronization and pull request data for the UI.
- **gitboard**: The command-line entrypoint for the application.
- **observability**: Manages OpenTelemetry and failure dump processing.
- **config**: Shared technical library for application configuration.

## System Flow
1. **Sync/Discovery**: `syncproj` and `remotegit` pull data from forges.
2. **Local Mapping**: `localgit` maps remote data to the user's local disk roots.
3. **Analysis**: `triage` and `pruneagent` process the combined data to find actionable insights.
4. **Delivery**: `server` and `dashboard` present the unified view to the user via a web interface.
