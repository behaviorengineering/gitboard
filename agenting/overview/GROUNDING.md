# Overview

  Gitboard is a local developer tool for monitoring GitHub/GitLab activity and local git state.

  ## System Summary
  The system operates by syncing remote forge data and local filesystem checkouts into a unified state. This state is then surfaced through a CLI and a web-based dashboard.

  ## Core Capabilities
  - **Discovery**: Automated syncing of GitHub orgs and GitLab groups.
  - **Monitoring**: Real-time visibility into CI status, MRs/PRs, and branch health.
  - **Automation**: AI-driven triage and branch pruning via the `pruneagent` and `triage` contexts.

  ## Architectural Guardrails
  - High-level orchestration resides in the `gitboard` slice.
  - Remote and local git operations are strictly separated into `remotegit` and `localgit`.
  - Configuration is centralized in the `config` slice.