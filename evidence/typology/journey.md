# Journey: gitboard Typology Refinement

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: cluster_proposal.md](cluster_proposal.md) · [Next: human_intervention.md](human_intervention.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

**Status**: Open

## Decisions Taken

- **Merge technical leaves into entrypoint**: Majordomo proposes merging `internal/observability`, `internal/server`, and `internal/syncproj` into the `gitboard` slice.
  - *Rejected*: Keeping them as separate slices.
  - *Reason*: They are technical leaves with single importers (the entrypoint) and serve bootstrapping or orchestration roles rather than independent domain logic.
- **Retain HTTP Surface as separate surface**: Majordomo rejected folding the `http_surface` role into the `entrypoint` package itself, but grouped the package under the `gitboard` slice to reflect ownership.
  - *Rejected*: Merging `internal/server` package into `cmd/gitboard`.
  - *Reason*: Maintains clear separation between the CLI entrypoint and the HTTP delivery mechanism.
- **Populated Business Objectives**: All slices were assigned concrete objectives to move from a "package inventory" to a functional catalog.
  - *Rejected*: Leaving objectives blank.
  - *Reason*: To resolve the "hollow slice" smell identified in validation.

## Technical Debt and Boundary Violations

| Finding | Smell | Alternatives | Lean |
|:---|:---|:---|:---|
| `gitboard` -> `gitboard-http` missing binding | Missing SliceBinding: `cmd/gitboard` imports `internal/server`. | 1. Add SliceBinding to catalog (Low cost). 2. Refactor to remove import (High cost). | Add SliceBinding to `refined_catalog_yaml`. |
| `gitboard-http` -> `dashboard` missing binding | Missing SliceBinding: `internal/server` imports `internal/dashboard`. | 1. Add SliceBinding to catalog (Low cost). 2. Refactor to remove import (High cost). | Add SliceBinding to `refined_catalog_yaml`. |
| `gitboard-http` -> `pruneagent` missing binding | Missing SliceBinding: `internal/server` imports `internal/pruneagent`. | 1. Add SliceBinding to catalog (Low cost). 2. Refactor to remove import (High cost). | Add SliceBinding to `refined_catalog_yaml`. |
| `gitboard-http` -> `triage` missing binding | Missing SliceBinding: `internal/server` imports `internal/triage`. | 1. Add SliceBinding to catalog (Low cost). 2. Refactor to remove import (High cost). | Add SliceBinding to `refined_catalog_yaml`. |
