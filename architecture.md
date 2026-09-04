# Architecture

  Gitboard is a Go-based application structured around several bounded contexts that manage the interaction between local git state, remote forge APIs, and the user interface.

  ## Core Components

  - **CLI &amp; Entrypoints**: The system provides a standard CLI (`gitboard`) and a TUI-driven service mode using `process-compose`.
  - **Dashboard**: A web-based UI served locally (default port `:1325`) that displays project status.
  - **Sync &amp; Discovery**: Logic to discover repositories across GitHub organizations and GitLab groups.
  - **Local Git Integration**: Interfaces with local disk roots to match `origin` remotes to tracked projects and worktrees.
  - **Triage &amp; LLM**: Optional AI-driven triage capabilities for analyzing repository data.
  - **Configuration**: A YAML-based configuration system located at `~/.config/gitboard/config.yaml`.

  ## System Topology
  The architecture follows a modular pattern where the `gitboard` command acts as a coordinator for specialized internal packages including `remotegit`, `localgit`, `server`, and `triage`.