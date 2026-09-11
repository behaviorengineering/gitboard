# Typology Cluster Proposal

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture_brief.md](architecture_brief.md) · [Next: journey.md](journey.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

  Majordomo proposes the following architecture-grounding refinements for the `gitboard` repository. These are proposals for the context branch and await human review.

  ## Proposed merges
  *No merges are recommended at this time.* While `graph_text` suggests merging some leaf packages into the entrypoint, Majordomo maintains that `internal/observability` and `internal/syncproj` are distinct technical components that should remain as independent packages to preserve modularity, even if they are currently door-private.

  ## Proposed renames
  *   `internal/remotegit` $\rightarrow$ `internal/forge_adapter` (or similar): Currently identified as an `adapter` via `package_roles`, but the name `remotegit` is slightly ambiguous regarding its role in filling `board` DTOs.
  *   `internal/localgit` $\rightarrow$ `internal/git_inspector`: To better reflect its role in inspecting local worktrees and being composed by the aggregator.

  ## Anti-pattern findings
  *   **Unclassified Domain Logic**: Several packages (`internal/localgit`, `internal/pruneagent`, `internal/remotegit`, `internal/syncproj`, `internal/triage`) are currently marked as `role: unknown` with `confidence: 0.00`. This indicates a lack of explicit role definition in the catalog that contradicts their high utility in the import graph.
  *   **Aggregator Ambiguity**: `internal/syncproj` is performing aggregation tasks (as seen in `package_roles`) but lacks a formal role assignment, creating a gap in the typology.

  ## Boundary debt
  *   **Unknown Role Debt**: `internal/localgit`, `internal/pruneagent`, `internal/remotegit`, `internal/syncproj`, and `internal/triage` all lack role definitions.
      *   *Smell*: High coupling and clear functional utility (e.g., `remotegit` fills `board` DTOs) without formal typology classification.
      *   *Alternatives*: 1. Manually assign roles (adapter, aggregator, etc.) in the catalog. 2. Leave as unknown.
      *   *Lean*: Majordomo recommends assigning roles based on observed `package_contracts`: `remotegit` as `adapter`, `syncproj` as `aggregator`, and `triage` as `aggregator`.
  *   **Documentation Gap**: Massive drift in `docs/develop/` paths across almost all internal packages.
      *   *Smell*: The `draft_catalog_yaml` and `architecture_draft` list numerous missing overview and component documentation paths.
      *   *Alternatives*: 1. Bulk generate stubs. 2. Ignore until feature-complete.
      *   *Lean*: Prioritize documentation for the `board` (DTO) and `cliexec` (runner) packages as they are core technical dependencies.

  ## Rationale
  The current topology shows a highly functional system where `cmd/gitboard` acts as a massive hub. However, the "unknown" status of 5 out of 13 packages prevents automated coverage checks and formal boundary enforcement. By promoting these from `unknown` to specific roles (adapter, aggregator, etc.), the Typology digest can begin enforcing capability constraints.

  ## Capability constraints
  ### is
  *   `cmd/gitboard`: `run_cli`
  *   `internal/board`: `data_shape`
  *   `internal/cliexec`: `exec_process`
  *   `internal/config`: `config`
  *   `internal/dashboard`: `aggregate_views`
  *   `internal/llm`: `adapt_external`, `fill_dto`
  *   `internal/observability`: `observability`
  *   `internal/remotegit`: `adapt_external`, `fill_dto`
  *   `internal/server`: `serve_http`, `wire_handlers`
  *   `internal/syncproj`: `aggregate_views`

  ### is-not
  *   `cmd/gitboard`: `own_domain_rules`, `serve_http`
  *   `internal/board`: `synchronize_state`, `merge_adapters`, `serve_http`, `orchestrate`, `own_domain_rules`, `run_cli`, `wire_handlers`
  *   `internal/cliexec`: `run_cli`, `own_domain_rules`
  *   `internal/config`: `own_domain_rules`
  *   `internal/dashboard`: `serve_http`, `run_cli`
  *   `internal/observability`: `config`, `own_domain_rules`
  *   `internal/server`: `own_domain_rules`

## Capability constraints (is / is-not)

Factual priors for refine. MUST NOT contradict.

- `cmd/gitboard` role=entrypoint is=[run_cli] must_not=[own_domain_rules, serve_http]
- `internal/board` role=dto is=[data_shape] must_not=[synchronize_state, merge_adapters, serve_http, orchestrate, own_domain_rules, run_cli, wire_handlers] filled_by=[internal/dashboard, internal/remotegit]
- `internal/cliexec` role=exec_runner is=[exec_process] must_not=[run_cli, own_domain_rules]
- `internal/config` role=config is=[config] must_not=[own_domain_rules]
- `internal/dashboard` role=aggregator is=[aggregate_views] must_not=[serve_http, run_cli]
- `internal/llm` role=adapter is=[adapt_external, fill_dto] must_not=[]
- `internal/localgit` role=unknown is=[] must_not=[]
- `internal/observability` role=observability is=[observability] must_not=[config, own_domain_rules]
- `internal/pruneagent` role=unknown is=[] must_not=[]
- `internal/remotegit` role=adapter is=[adapt_external, fill_dto] must_not=[]
- `internal/server` role=server is=[serve_http, wire_handlers] must_not=[own_domain_rules]
- `internal/syncproj` role=aggregator is=[aggregate_views] must_not=[serve_http, run_cli]
- `internal/triage` role=unknown is=[] must_not=[]

## Mechanical override (Majordomo)

Rejected folding server into entrypoint. Keep delivery on its own slice/surface.

- MUST NOT merge `internal/server` (server) into an entrypoint package.
- Entrypoint `cmd/gitboard` may import HTTP packages as wiring only.
