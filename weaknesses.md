# Weaknesses

### Architectural Boundary Violations
* **Unsanctioned Dashboard Coupling**: `internal/dashboard` (UI) has undocumented dependencies on `config`, `gitboard` (domain), and `remotegit`.
* **CLI Orchestration Drift**: `cmd/gitboard` (CLI runner) bypasses declared bindings to interact directly with `localgit` and `remotegit`.
* **Sync Projection Leakage**: `internal/syncproj` (gitboard component) imports `remotegit` without a formal slice binding.
* **Remotegit Domain Leakage**: `internal/remotegit` reaches back into `config` and `gitboard` (domain), creating circularity risks.
* **Agentic Workflow Violations**: `internal/pruneagent` (triage component) directly imports `localgit` and `remotegit` without declared bindings.