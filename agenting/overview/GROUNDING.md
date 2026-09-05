# Overview

  Gitboard is a local developer dashboard for GitHub and GitLab.

  ## Mission
  To provide visibility into CI status, MRs/PRs, and local git states using existing CLI tools.

  ## Architecture Summary
  The system is split into a primary `gitboard` orchestration slice and a `config` management slice. The `gitboard` slice encompasses the CLI, API, and UI surfaces, alongside domain logic for git operations, AI triage, and pruning agents. Current architectural focus involves managing the boundary debt between functional agentic roles and the core orchestration domain.