# Weaknesses

  * **Undocumented Infrastructure Coupling**: Multiple slices (`gitboard`, `pruneagent`, `triage`) rely on `internal-config` and `internal-llm` (owned by `board`), but these SliceBindings are missing.
  * **Orchestration Dependency Leak**: `localgit`, `pruneagent`, and `remotegit` depend on `internal-cliexec` (owned by `gitboard`) without formal SliceBinding declarations.