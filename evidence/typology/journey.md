# Journey: gitboard Typology Refinement

**Status:** Open

## Decisions Taken

- **Consolidated Orchestration:** Merged `observability`, `server`, and `syncproj` into the `gitboard` slice.
  - *Rejected:* Keeping them as standalone slices. This would create a fragmented architecture for a single-binary tool.
  - *Lean:* Unify application logic under the primary entrypoint.
- **Demoted Exec Adapter:** Moved `cliexec` from `kind: cli` to `gitboard` owns.
  - *Rejected:* Maintaining `cliexec` as a peer `kind: cli` slice. This violated the rule against non-main CLI surfaces.
  - *Lean:* Treat as a technical utility owned by the main application.
- **Reclassified Infrastructure:** Moved `config` and `llm` to `libraries[]`.
  - *Rejected:* Maintaining them as peer domain slices. This creates high cognitive load by mixing infra with domain logic.
  - *Lean:* Use technical libraries to distinguish infrastructure from business logic.

## Technical Debt and Boundary Violations

| Finding | Smell | Alternatives | Lean |
| :--- | :--- | :--- | :--- |
| `gitboard` (cmd-gitboard -> internal-config) | Missing SliceBinding to `board` | 1. Add SliceBinding (Lean). 2. Refactor config out of `board` (High cost). | Approve binding to `board` library. |
| `gitboard` (cmd-gitboard -> internal-llm) | Missing SliceBinding to `board` | 1. Add SliceBinding (Lean). 2. Move LLM to a peer slice (High cost). | Approve binding to `board` library. |
| `gitboard` (internal-server -> internal-config) | Missing SliceBinding to `board` | 1. Add SliceBinding (Lean). 2. Move server to a different slice (High cost). | Approve binding to `board` library. |
| `gitboard` (internal-syncproj -> internal-config) | Missing SliceBinding to `board` | 1. Add SliceBinding (Lean). 2. Move syncproj to a different slice (High cost). | Approve binding to `board` library. |
| `localgit` (internal-localgit -> internal-cliexec) | Missing SliceBinding to `gitboard` | 1. Add SliceBinding (Lean). 2. Move cliexec to a shared library (Medium cost). | Approve binding to `gitboard` slice. |
| `pruneagent` (internal-pruneagent -> internal-llm) | Missing SliceBinding to `board` | 1. Add SliceBinding (Lean). 2. Promote LLM to a domain slice (High cost). | Approve binding to `board` library. |
| `pruneagent` (internal-pruneagent -> internal-cliexec) | Missing SliceBinding to `gitboard` | 1. Add SliceBinding (Lean). 2. Move cliexec to a shared library (Medium cost). | Approve binding to `gitboard` slice. |
| `remotegit` (internal-remotegit -> internal-cliexec) | Missing SliceBinding to `gitboard` | 1. Add SliceBinding (Lean). 2. Move cliexec to a shared library (Medium cost). | Approve binding to `gitboard` slice. |
| `triage` (internal-triage -> internal-config) | Missing SliceBinding to `board` | 1. Add SliceBinding (Lean). 2. Move config to a global library (Medium cost). | Approve binding to `board` library. |
| `triage` (internal-triage -> internal-llm) | Missing SliceBinding to `board` | 1. Add SliceBinding (Lean). 2. Promote LLM to a domain slice (High cost). | Approve binding to `board` library. |