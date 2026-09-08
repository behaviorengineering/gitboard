# Weaknesses

### Boundary Debt
- **Unrecorded Git Couplings**: `internal-remotegit` currently depends on `board` and `config` without formal SliceBindings.
- **CLI-to-Utility Gap**: The `gitboard` CLI utilizes `internal-cliexec` (exec-utils) without a formal architectural binding.
- **LLM Config Dependency**: `internal-llm` relies on `internal-config` without a formal binding.

### Structural Risks
- **Platform Hub Risk**: The `config` package acts as a high-degree dependency for nearly all slices, creating a central point of coupling.
- **Dashboard Fragmentation**: The `dashboard` slice currently acts as an intermediary between the `server` and `board`, suggesting a potential need to consolidate dashboard logic into the server slice.