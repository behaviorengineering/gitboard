# Typology Cluster Proposal: gitboard

## Proposed merges

### Merge `internal/board` into `internal/dashboard`
- **Why**: `internal/board` is a DTO-only package (`deliveryHint: dto`) containing JSON data types shared across the dashboard. It lacks independent domain logic and serves as the data contract for the UI.
- **Rejected**: Keeping `board` as a peer slice. This creates unnecessary fragmentation for a package that only exports data structures.
- **Cost of alternative**: Higher cognitive load to track data models across two slices; increased boilerplate for cross-slice imports.
- **Lean**: Fold the DTOs into the consumer slice (`dashboard`) to unify the domain model with its primary surface.

### Merge `internal/cliexec` into `internal/localgit` or `internal/remotegit` (as a library)
- **Why**: `cliexec` is an exec-adapter (exports `Runner`) used by git-related packages to run shell commands. It is not a CLI surface itself.
- **Rejected**: Keeping `cliexec` as a standalone `kind: cli` slice.
- **Cost of alternative**: A "platform" or "cli" slice that only contains a single utility runner is a hollow abstraction.
- **Lean**: Reclassify `cliexec` as a library or move it into a specialized `internal/git/exec` utility package.

## Proposed renames

### Rename `internal/server` to `internal/dashboard/server`
- **Why**: `internal/server` is an `http-surface` that hosts the dashboard UI. It is currently a top-level slice but functions as the delivery mechanism for the dashboard domain.
- **Rejected**: Keeping it as a top-level `server` slice.
- **Cost of alternative**: Creates a "platform" slice that is actually just a domain-specific web server.
- **Lean**: Nest the server logic within the `dashboard` context to reflect its role as the dashboard's HTTP surface.

## Anti-pattern findings

### Capability-as-Domain Pillar
- **Finding**: `internal/llm` is currently treated as a peer domain slice.
- **Smell**: LLM integration is a capability/utility used by `triage` and `pruneagent`. It is not a bounded context of the product's core business logic (which is git/board management).
- **Recommendation**: Reclassify `llm` as a library or a component within the `triage` slice.

### Temporal/Pipeline Slices
- **Finding**: `internal/syncproj` and `internal/triage` are structured as peer slices.
- **Smell**: These appear to be stages in a data pipeline (Sync -> Triage -> Dashboard).
- **Recommendation**: These should be viewed as internal processes of the `gitboard` domain rather than independent bounded contexts.

## Boundary debt

| Smell | Alternatives | Lean |
|-------|--------------|------|
| **Orphaned Observability**: `internal/observability` is only imported by `cmd/gitboard`. | 1. Keep as a library. 2. Fold into `cmd/gitboard`. | Move to `libraries/observability` to signal it is a technical utility, not a domain slice. |
| **DTO Fragmentation**: `internal/board` exists as a separate slice for JSON tags. | 1. Keep as is. 2. Move to `internal/dashboard`. | Move to `internal/dashboard` to consolidate the domain model. |
| **High-Coupling Hub**: `cmd/gitboard` imports 11 different packages. | 1. Refactor to a thin entry point. 2. Accept as a "composition root". | Accept as a composition root, but ensure all logic remains in `internal/` to prevent the CLI from becoming a "god package". |

## Rationale

Majordomo proposes this consolidation to move `gitboard` from a collection of loosely coupled packages toward a cohesive set of bounded contexts. The current structure suffers from "package sprawl," where technical utilities (like `cliexec` and `observability`) and data contracts (like `board`) are promoted to the same status as primary domain drivers (like `localgit`). 

By grouping the HTTP surface with the `dashboard` and treating the `llm` as a capability rather than a pillar, we reduce the complexity of the import graph and clarify the product's actual purpose: managing git state via a visual board. The leanest path is to consolidate the "delivery" components (HTTP/CLI) closer to the domain logic they serve.