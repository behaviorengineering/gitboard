# Weaknesses

- **Unbounded Cross-Slice Imports**: Multiple slices (`dashboard`, `git`, `triage`) are importing `config` without formal `sliceBindings`.
- **Leaky Git Abstractions**: `remotegit` has unexpected dependencies on `board` and `cliexec`, violating the intended `git` slice boundaries.
- **Implicit Capability Dependencies**: `pruneagent` and `triage` rely on `llm` (via `gitboard`) without explicit architectural declarations.
- **Orchestrator Bloat**: `gitboard` (cmd) is directly importing `cliexec`, bypassing potential intermediate layers.