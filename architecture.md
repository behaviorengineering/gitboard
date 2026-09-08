# Architecture

> **Teaching story.** Living project architecture for humans and review grounding. Not Typology seed evidence (see `evidence/typology/architecture_brief.md`).

Gitboard is organized into functional slices that separate command-line orchestration, web visualization, and domain logic.

## Bounded Contexts

### Core Orchestration
- **gitboard**: The primary CLI entry point. It orchestrates core operations and manages the lifecycle of the application.
- **exec-utils**: Provides low-level execution primitives and runners used by the CLI and other services to process commands.

### Domain &amp; Data
- **board**: Manages the core domain logic for project state and the representation of the developer's board.
- **git**: Handles all interactions with local and remote git repositories and manages state synchronization.
- **config**: A centralized platform utility providing configuration management across the entire system.

### Intelligence &amp; Analysis
- **llm**: Integrates large language models to provide intelligent analysis and triage capabilities.
- **triage**: Analyzes project data to categorize and prioritize maintenance actions.
- **pruneagent**: Automates the identification and execution of project cleanup tasks.

### Visualization
- **dashboard**: Serves as the web-based visualization layer, presenting project and board data to the user via a browser.

## System Flow
The `gitboard` CLI triggers synchronization and orchestration, which utilizes `git` and `config` to gather data. This data is processed through `board` and `triage` logic, eventually being served to the `dashboard` for user interaction.
