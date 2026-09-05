## Proposed merges
- **gitboard** slice: merge `internal/observability`, `internal/server`, and `internal/syncproj` into this slice.
- **git** slice: merge `internal/localgit` and `internal/remotegit` into a single slice named `git`.

## Proposed renames
- No rename needed; slice names already reflect domain concerns.

## Anti‑pattern findings
- No temporal pipeline stages are defined as peer slices.
- The `llm` slice is a capability slice and is not promoted to a domain pillar.
- CLI surfaces (`cliexec-cli`, `gitboard-cli`) are correctly placed under their respective domain slices.
- No projection‑as‑slice peers are present.

## Boundary debt
| Slice | Debt | Action |
|-------|------|--------|
| `gitboard` | `observability`, `server`, `syncproj` are sole importers → merge | Merge into `gitboard` |
| `git` | `localgit` and `remotegit` are separate but share a stem | Merge into `git` |
| `gitboard` | Imports many slices; consider extracting shared infrastructure | Create a shared `git` infrastructure slice if needed |

## Rationale
The merge heuristics identify packages with a single importer (`observability`, `server`, `syncproj`) as candidates to be colocated with the importer (`gitboard`). This reduces coupling and simplifies the slice boundary. Merging `localgit` and `remotegit` into a single `git` slice aligns with the same‑job‑family heuristic and removes duplicate infrastructure concerns. All anti‑patterns are respected: CLI surfaces remain within domain slices, capabilities stay separate, and no temporal pipeline stages are exposed as peers. The boundary debt table records the necessary changes to achieve a cleaner, more coherent architecture.