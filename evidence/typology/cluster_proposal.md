# Proposed merges

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: architecture_brief.md](architecture_brief.md) · [Next: journey.md](journey.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

Majordomo proposes the following merges based on observed role companionship and sole-importer heuristics:
- `internal/remotegit` + `internal/localgit` $\rightarrow$ `internal/git` (Proposed): Both are adapters handling git-related external data. While currently separate, they share a functional stem.
- `internal/triage` + `internal/pruneagent` $\rightarrow$ `internal/analysis` (Proposed): Both are aggregators/unknowns focused on branch/worktree health.

# Proposed renames
- `internal/board` $\rightarrow$ `internal/dto/board`: To explicitly signal its role as a technical data-shape library rather than a domain slice.
- `internal/cliexec` $\rightarrow$ `internal/runner`: To align with its `exec_runner` role.

# Anti-pattern findings
- **Capability Promotion**: `internal/dashboard` and `internal/syncproj` are currently categorized as slices in the draft, but they function as aggregators. They should be treated as technical orchestration layers rather than independent domain pillars.
- **Role Ambiguity**: `internal/localgit`, `internal/remotegit`, `internal/pruneagent`, `internal/syncproj`, and `internal/triage` all have `confidence: 0.00`. This creates a "dark matter" effect in the topology where the core logic is unclassified.

# Boundary debt
- **Aggregator/Domain Blur**: `internal/dashboard` composes `localgit` and `remotegit`. There is a smell of "God Aggregator" where the dashboard is becoming the de facto owner of the orchestration logic. 
  - *Alternative*: Move orchestration logic into a dedicated `internal/orchestrator` slice.
  - *Lean*: Explicitly classify `dashboard` as an aggregator and accept the coupling as part of the product slice.
- **Unreached/Unknown Packages**: The high number of `unknown` roles (5 packages) prevents a clean slice-to-library mapping.
  - *Alternative*: Perform a deep AST scan to assign `adapter` or `aggregator` roles.
  - *Lean*: Accept the debt in this pass and flag for next Typology cycle.

# Rationale
The proposal prioritizes stabilizing the technical foundation (libraries/dto) before attempting to define product slices. By grouping the git-related adapters and the analysis-related aggregators, we reduce the cognitive load of the 13-package sprawl. The distinction between `cmd/gitboard` (entrypoint) and `internal/server` (server) is strictly maintained to respect the two distinct delivery doors.

# Capability constraints
### cmd/gitboard
- **is**: `run_cli`, `orchestrate`
- **must_not**: `own_domain_rules`, `serve_http`, `exec_process`, `fill_dto`, `adapt_external`, `aggregate_views`, `data_shape`, `observability`, `config`, `wire_handlers`, `synchronize_state`, `merge_adapters`

### internal/board
- **is**: `data_shape`
- **must_not**: `synchronize_state`, `merge_adapters`, `serve_http`, `orchestrate`, `own_domain_rules`, `run_cli`, `wire_handlers`, `exec_process`, `fill_dto`, `adapt_external`, `aggregate_views`, `observability`, `config`

### internal/cliexec
- **is**: `exec_process`
- **must_not**: `run_cli`, `own_domain_rules`, `orchestrate`, `fill_dto`, `adapt_external`, `aggregate_views`, `data_shape`, `observability`, `config`, `serve_http`, `wire_handlers`, `synchronize_state`, `merge_adapters`

### internal/config
- **is**: `config`
- **must_not**: `own_domain_rules`, `orchestrate`, `exec_process`, `fill_dto`, `adapt_external`, `aggregate_views`, `data_shape`, `observability`, `serve_http`, `wire_handlers`, `run_cli`, `synchronize_state`, `merge_adapters`

### internal/dashboard
- **is**: `aggregate_views`
- **must_not**: `serve_http`, `run_cli`, `orchestrate`, `exec_process`, `adapt_external`, `data_shape`, `observability`, `config`, `synchronize_state`, `merge_adapters`, `wire_handlers`

### internal/llm
- **is**: `adapt_external`, `fill_dto`
- **must_not**: `orchestrate`, `exec_process`, `own_domain_rules`, `aggregate_views`, `data_shape`, `observability`, `config`, `run_cli`, `synchronize_state`, `merge_adapters`

### internal/localgit
- **is**: `adapt_external`, `fill_dto`
- **must_not**: `orchestrate`, `exec_process`, `own_domain_rules`, `aggregate_views`, `data_shape`, `observability`, `config`, `serve_http`, `wire_handlers`, `run_cli`, `synchronize_state`, `merge_adapters`

### internal/observability
- **is**: `observability`
- **must_not**: `config`, `own_domain_rules`, `orchestrate`, `exec_process`, `fill_dto`, `adapt_external`, `aggregate_views`, `data_shape`, `serve_http`, `wire_handlers`, `run_cli`, `synchronize_state`, `merge_adapters`

### internal/remotegit
- **is**: `adapt_external`, `fill_dto`
- **must_not**: `orchestrate`, `exec_process`, `own_domain_rules`, `aggregate_views`, `data_shape`, `observability`, `config`, `serve_http`, `wire_handlers`, `run_cli`, `synchronize_state`, `merge_adapters`

### internal/server
- **is**: `serve_http`, `wire_handlers`
- **must_not**: `own_domain_rules`, `orchestrate`, `exec_process`, `fill_dto`, `adapt_external`, `aggregate_views`, `data_shape`, `observability`, `config`, `run_cli`, `synchronize_state`, `merge_adapters`

## Capability constraints (is / is-not)

Factual priors for refine. MUST NOT contradict.

- `cmd/gitboard` role=entrypoint is=[run_cli, orchestrate] must_not=[own_domain_rules, serve_http, exec_process, fill_dto, adapt_external, aggregate_views, data_shape, observability, config, wire_handlers, synchronize_state, merge_adapters]
- `internal/board` role=dto is=[data_shape] must_not=[synchronize_state, merge_adapters, serve_http, orchestrate, own_domain_rules, run_cli, wire_handlers, exec_process, fill_dto, adapt_external, aggregate_views, observability, config] filled_by=[internal/dashboard, internal/remotegit]
- `internal/cliexec` role=exec_runner is=[exec_process] must_not=[run_cli, own_domain_rules, orchestrate, fill_dto, adapt_external, aggregate_views, data_shape, observability, config, serve_http, wire_handlers, synchronize_state, merge_adapters]
- `internal/config` role=config is=[config] must_not=[own_domain_rules, orchestrate, exec_process, fill_dto, adapt_external, aggregate_views, data_shape, observability, serve_http, wire_handlers, run_cli, synchronize_state, merge_adapters]
- `internal/dashboard` role=aggregator is=[aggregate_views] must_not=[serve_http, run_cli, orchestrate, exec_process, adapt_external, data_shape, observability, config, synchronize_state, merge_adapters, wire_handlers]
- `internal/llm` role=adapter is=[adapt_external, fill_dto] must_not=[orchestrate, exec_process, own_domain_rules, aggregate_views, data_shape, observability, config, run_cli, synchronize_state, merge_adapters]
- `internal/localgit` role=adapter is=[adapt_external, fill_dto] must_not=[orchestrate, exec_process, own_domain_rules, aggregate_views, data_shape, observability, config, serve_http, wire_handlers, run_cli, synchronize_state, merge_adapters]
- `internal/observability` role=observability is=[observability] must_not=[config, own_domain_rules, orchestrate, exec_process, fill_dto, adapt_external, aggregate_views, data_shape, serve_http, wire_handlers, run_cli, synchronize_state, merge_adapters]
- `internal/pruneagent` role=unknown is=[] must_not=[orchestrate, exec_process, fill_dto, own_domain_rules, adapt_external, aggregate_views, data_shape, observability, config, serve_http, wire_handlers, run_cli, synchronize_state, merge_adapters]
- `internal/remotegit` role=adapter is=[adapt_external, fill_dto] must_not=[orchestrate, exec_process, own_domain_rules, aggregate_views, data_shape, observability, config, serve_http, wire_handlers, run_cli, synchronize_state, merge_adapters]
- `internal/server` role=server is=[serve_http, wire_handlers] must_not=[own_domain_rules, orchestrate, exec_process, fill_dto, adapt_external, aggregate_views, data_shape, observability, config, run_cli, synchronize_state, merge_adapters]
- `internal/syncproj` role=aggregator is=[aggregate_views] must_not=[serve_http, run_cli, orchestrate, exec_process, fill_dto, adapt_external, data_shape, observability, config, synchronize_state, merge_adapters, wire_handlers]
- `internal/triage` role=aggregator is=[aggregate_views] must_not=[serve_http, run_cli, orchestrate, exec_process, fill_dto, adapt_external, data_shape, observability, config, synchronize_state, merge_adapters, wire_handlers]

## Mechanical override (Majordomo)

Rejected folding server into entrypoint. Keep delivery on its own slice/surface.

- MUST NOT merge `internal/server` (server) into an entrypoint package.
- Entrypoint `cmd/gitboard` may import HTTP packages as wiring only.
