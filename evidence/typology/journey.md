# Journey: Gitboard Typology Refinement

## Status
Open (Architecture findings pending human decision)

## Decisions
- **Consolidated Entry Point**: Merged `internal/observability`, `internal/server`, and `internal/syncproj` into the `gitboard` slice as they are sole-importer leaves serving the main CLI.
- **Domain Consolidation**: Merged `llm`, `pruneagent`, and `triage` into a single `triage` domain to reflect their shared agentic workflow purpose.
- **Demoted Exec Adapter**: Moved `internal/cliexec` from a standalone slice to a component owned by `remotegit`, as it is an execution helper without its own `main`.
- **Substantive Objectives**: Replaced all hollow/empty objectives with concrete business value statements.
- **Surface Realignment**: Moved `internal/dashboard` under a `ui` surface.

## Technical Debt and Boundary Violations

| Package | Violation | Mitigation |
|---------|-----------|------------|
| `internal/board` | Missing documentation paths in draft | Manual update of docs/develop/board/ files |
| `internal/remotegit` | High coupling to `cliexec` | Refactor to use interface-based execution if needed |
| `internal/dashboard` | Missing SliceBinding to `config` | Approve binding or refactor dependency |
| `internal/dashboard` | Missing SliceBinding to `gitboard` | Approve binding or refactor dependency |
| `internal/dashboard` | Missing SliceBinding to `remotegit` | Approve binding or refactor dependency |
| `cmd/gitboard` | Missing SliceBinding to `localgit` | Approve binding or refactor dependency |
| `cmd/gitboard` | Missing SliceBinding to `remotegit` (via `cliexec`) | Approve binding or refactor dependency |
| `cmd/gitboard` | Missing SliceBinding to `remotegit` (via `remotegit`) | Approve binding or refactor dependency |
| `internal/syncproj` | Missing SliceBinding to `remotegit` | Approve binding or refactor dependency |
| `internal/remotegit` | Missing SliceBinding to `config` | Approve binding or refactor dependency |
| `internal/remotegit` | Missing SliceBinding to `gitboard` | Approve binding or refactor dependency |
| `internal/pruneagent` | Missing SliceBinding to `localgit` | Approve binding or refactor dependency |
| `internal/pruneagent` | Missing SliceBinding to `remotegit` (via `cliexec`) | Approve binding or refactor dependency |