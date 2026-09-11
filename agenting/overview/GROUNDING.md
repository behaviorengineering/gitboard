# Overview

High-level project grounding for review.

## Mission Summary
Gitboard is a local dashboard for managing GitHub/GitLab forge data and local git state, focusing on CI status, PRs, and branch lifecycle.

## Architecture Summary
The system is composed of specialized slices:
- **Data Ingestion**: `remotegit`, `syncproj`.
- **Local Inspection**: `localgit`, `cliexec`.
- **Intelligence/Triage**: `triage`, `pruneagent`, `llm`.
- **Presentation**: `server`, `dashboard`, `gitboard`.
- **Core Domain**: `board`.

This architecture aims to decouple the source of truth (forges/local disk) from the presentation layer via internal DTOs.