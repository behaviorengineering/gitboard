# Overview

  Gitboard is a local developer dashboard for monitoring GitHub and GitLab activity.

  ## Mission Summary
  To provide a unified local interface for CI status, MR/PR management, and local git state to reduce context switching.

  ## Architectural Summary
  The system follows a bounded-context pattern:
  - **Orchestration**: `gitboard` slice manages the lifecycle and CLI.
  - **Data/Logic**: `board` (domain) and `git-engine` (git operations) provide the core functionality.
  - **Configuration**: `config` serves as a shared platform leaf.
  - **Capabilities**: `gitboard-llm-gateway` provides AI-driven triage.
  - **Presentation**: `dashboard` provides the visual monitoring layer.