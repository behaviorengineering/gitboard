<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: mission.md](mission.md) · [Next: conventions.md](conventions.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

markdown: |
# Architecture
> Teaching story. Living project architecture for humans and review grounding. Not Typology seed evidence (see `evidence/typology/architecture_brief.md`).
The system operates as an orchestration layer for Git-based workflows, bridging remote forge data with intelligent analysis. The architecture follows a flow from discovery to actionable intelligence:
### Discovery and Orchestration
At the edge, `gitboard` serves as the primary command-line entrypoint, orchestrating the lifecycle of the application. Project discovery and selection are handled by `syncproj`, which aggregates views to provide a coherent interface for navigating repositories.
### Integration and Connectivity
The system interacts with the external world through `git-adapters`, providing standardized clients for various Git forge APIs like GitHub and GitLab. Communication with intelligence models is abstracted through the `llm` package, which provides an OpenAI-compatible interface for all reasoning tasks.
### Analysis and Intelligence
The core value is delivered through the `triage` package, which analyzes project and job data to produce intelligent responses. Supporting this is `pruneagent`, which investigates local or removable branches to maintain repository health.
### Infrastructure and Observability
The system is exposed via a `server` package providing HTTP interfaces and mux routing. Reliability is maintained through `observability`, which bootstraps OpenTelemetry (OTEL) and manages error-only inference dumps for debugging complex LLM interactions.
