# Overview

High-level project grounding for review.

## Mission Summary
Gitboard is a local-first dashboard for monitoring GitHub and GitLab activity (CI, MRs, triage) using local forge CLIs and optional LLM assistance.

## Architectural Summary
The system is organized into functional slices:
- **Orchestration**: `gitboard` (CLI).
- **Data**: `dto/board` (shared types), `git` (forge adapters), `runner` (process execution).
- **Intelligence**: `analysis` (triage/pruning), `llm` (AI interface).
- **UI**: `dashboard` (aggregator), `server` (HTTP/Web UI), `syncproj` (discovery).
- **Infrastructure**: `config` (settings), `observability` (telemetry).

**Key Boundary Note**: The `analysis` slice currently has undeclared dependencies on `git` and `runner` which are noted as technical debt.