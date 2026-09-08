# Typology Journey: gitboard

## Status
Refined via Majordomo Typology digest. Merges and role reclassifications applied based on cluster proposal and package contract evidence.

## Decisions

| Decision | Rejected | Reason |
| :--- | :--- | :--- |
| **Merge `observability`, `server`, `syncproj` into `gitboard`** | Keep as standalone slices | They are sole-importers of the entrypoint or serve as technical surfaces for the CLI, causing unnecessary slice sprawl. |
| **Classify `localgit`/`remotegit` as adapters** | Keep as domain slices | Evidence shows they are technical interfaces to external tools (git/APIs) rather than independent business domains. |
| **Classify `triage` as DTO slice** | Keep as domain slice | Package contracts show it primarily exports request/response structures for analysis. |
| **Classify `pruneagent` as runner** | Keep as domain slice | Evidence shows it is an orchestration layer for executing `cliexec` tasks. |

## Technical Debt and Boundary Violations

| Smell | Alternatives | Lean |
| :--- | :--- | :--- |
| **Entrypoint Bloat**: `cmd/gitboard` now owns `server` and `observability`. | 1. Move to `libraries/` (Medium cost). 2. Accept as surface container (Low cost). | Accept the entrypoint as the primary surface container for this small-scale tool. |
| **Role Ambiguity**: `localgit` and `remotegit` remain in `internal/` but act as adapters. | 1. Move to `pkg/adapters/` (High cost). 2. Explicitly classify in catalog (Low cost). | Explicitly classify as adapters in the next catalog iteration. |