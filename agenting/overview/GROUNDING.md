# Overview

# Mission

Gitboard exists to provide developers with a unified, local dashboard for managing their workflow across GitHub and GitLab. 

By aggregating CI status, pull/merge requests, and local git state, it reduces context-switching and helps developers triage failed jobs and manage local checkouts directly from a single interface.

## Architecture

# Architecture

> **Teaching story.** This document describes the high-level shape of the system.

Gitboard is organized into functional slices that separate domain logic from external interfaces (adapters) and technical surfaces.

## Bounded Contexts

### Core Orchestration &amp; UI
- **Dashboard**: The central orchestrator that collects data from various adapters to provide a unified project view.
- **Gitboard**: The primary entrypoint, managing the CLI interface and the local dashboard server. It contains technical surfaces like `observability` and `server`.
- **Board**: Provides the shared data structures used to maintain dashboard state and synchronize forge data.

### Adapters (External Interfaces)
- **LocalGit**: An adapter for performing local filesystem operations and git inspections.
- **RemoteGit**: An adapter for interacting with GitHub and GitLab remote APIs.
- **LLM**: An OpenAI-compatible client used for automated code and triage analysis.

### Specialized Logic
- **Triage**: Provides the logic and DTOs required to analyze and summarize triage requests.
- **PruneAgent**: An orchestration layer that executes automated investigations into removable local branches.
- **Cliexec**: A standardized technical interface for executing system commands and shell processes.

### Libraries
- **Config**: A shared technical package for application configuration management.

## Boundary Notes
The system treats `localgit` and `remotegit` as technical adapters to external tools rather than independent business domains. The entrypoint (`gitboard`) acts as a container for the server and observability components to minimize slice sprawl.
