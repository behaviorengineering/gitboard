# Weaknesses

  ## Technical Debt
  - **Capability Leak**: `llm` is currently a standalone package rather than a capability integrated into the `triage` and `pruneagent` domains.
  - **Fragmented Git Domain**: Physical package paths (`localgit` and `remotegit`) do not yet match the unified `internal/git` typology.
  - **Surface/Domain Confusion**: `cliexec` is currently implemented as a package rather than being strictly defined as a CLI surface component.

  ## Architectural Drift
  - Cross-slice imports exist between `cmd-gitboard`/`internal-server` and `pruneagent`/`triage` that are not yet formally declared in slice bindings.