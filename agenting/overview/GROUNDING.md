# Overview

  Gitboard is a local developer tool for monitoring GitHub/GitLab activity and local git state via a web dashboard. 

  **Key Architecture Pillars:**
  - **Local-First**: Relies on local `gh`, `glab`, and `git` installations.
  - **Modular Internal Design**: Uses bounded contexts for remote git interaction, local git management, and UI serving.
  - **Config-Driven**: Behavior is controlled via a centralized YAML configuration.

  **Core Mission:**
  To provide a unified, local-running dashboard for CI status, MR/PR tracking, and job triage.