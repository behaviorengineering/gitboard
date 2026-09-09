# Typology Cluster Proposal: gitboard

## Proposed merges

Majordomo proposes the following merges to consolidate technical leaves and reduce surface area. These are architecture-grounding proposals for the context branch.

- **Consolidate Observability**: Merge `internal/observability` into `cmd/gitboard`. 
  - *Reasoning*: `internal/observability` is a leaf package with a sole importer (`cmd/gitboard`). It serves a technical bootstrapping role for the entrypoint.
- **Consolidate Server/UI**: Merge `internal/server` into `cmd/gitboard`.
  - *Reasoning*: `internal/server` is a sole importer of the entrypoint's scope and functions as the HTTP surface for the CLI-driven application.
- **Consolidate Sync Logic**: Merge `internal/syncproj` into `cmd/gitboard`.
  - *Reasoning*: `internal/syncproj` is a leaf package with a sole importer (`cmd/gitboard`).

## Proposed renames

- **Standardize DTO Naming**: No renames proposed; `internal/board` correctly identifies as a `dto` role via AST evidence.

## Anti-pattern findings

- **Role/Folder Confusion**: No significant role mismatches detected. `internal/cliexec` is correctly identified as `exec_runner` (imports `os/exec`) rather than being mislabeled as a domain package or a CLI surface.
- **Capability Promotion**: No domain pillars were incorrectly promoted from technical adapters.

## Boundary debt

The following items represent documentation and metadata gaps that create "hollow" slices in the current catalog.

- **Missing Domain Objectives**: Most slices (`board`, `cliexec`, `dashboard`, `llm`, `localgit`, `observability`, `pruneagent`, `remotegit`, `server`, `syncproj`, `triage`) lack a defined `objective` in the `draft_catalog_yaml`.
  - *Smell*: The catalog describes *what* the packages are but not *why* they exist in the product context.
  - *Alternatives*: 1. Leave as is (Low cost, high debt). 2. Populate objectives via product discovery (Medium cost, low debt).
  - *Lean*: Populate objectives during the next documentation sprint.
- **Documentation Path Mismatches**: Multiple packages (e.g., `internal/board`, `internal/cliexec`, `internal/dashboard`) have missing or incomplete `docs/develop/` paths in the draft catalog.
  - *Smell*: Disconnect between the code structure and the documentation manifest.
  - *Alternatives*: 1. Update catalog to match existing files (Low cost). 2. Create missing files (High cost).
  - *Lean*: Update `draft_catalog_yaml` to reflect actual file locations to ensure coverage checks pass.

## Rationale

This proposal prioritizes the reduction of "leaf noise" by moving technical utilities (`observability`, `syncproj`) and the primary HTTP surface (`server`) into the entrypoint's immediate orbit. This recognizes that `cmd/gitboard` acts as the orchestrator for the entire local toolset. The boundary debt focuses on the transition from a "package inventory" to a "functional catalog" by addressing the lack of semantic objectives.

## Mechanical override (Majordomo)

Rejected folding http_surface into entrypoint. Keep HTTP delivery on its own slice/surface.

- MUST NOT merge `internal/server` (http_surface) into an entrypoint package.
- Entrypoint `cmd/gitboard` may import HTTP packages as wiring only.
