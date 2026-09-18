# Architecture

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: mission.md](mission.md) · [Next: conventions.md](conventions.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

> Teaching story. Living project architecture for humans and review grounding. Not Typology seed evidence (see `evidence/typology/architecture_brief.md`).
The `gitboard` system orchestrates local code-change board operations through a command-line entrypoint. The architecture is composed of several functional slices that manage git operations, project discovery, and data presentation.
At its core, the system handles repository interaction through `localgit`, which provides git operations and metadata enrichment for local repositories, and `remotegit`, which provides adapters to interact with remote git services such as GitHub and GitLab. Project scope is managed by `syncproj`, which aggregates project selection and discovery information.
The system facilitates analysis and maintenance via `triage`, which analyzes project logs and job information to produce a summary and root cause analysis, and `pruneagent`, which investigates local-only or likely-removable branches via the `Service.Investigate` method. Execution of system tasks is handled by `cliexec`, which executes external processes using the `os/exec` package.
Data and observability are integrated through several components. The `board` slice provides JSON data structures used by the dashboard and other components, while `dashboard` aggregates views for the dashboard. The `server` slice provides HTTP serving capabilities and multiplexer initialization. For intelligence and monitoring, `llm` provides an OpenAI-compatible chat completions client by adapting external API responses, and `observability` initializes OpenTelemetry tracing and manages error-only inference dumps.

## Grounded slice objectives

- **board**: Provides JSON data structures used by the dashboard and other components.
- **cliexec**: Executes external processes using the os/exec package.
- **dashboard**: Aggregates views for the dashboard.
- **gitboard**: The package serves as a command-line entrypoint that orchestrates local code-change board operations.
- **llm**: Provides an OpenAI-compatible chat completions client by adapting external API responses.
- **localgit**: Provides git operations and metadata enrichment for local repositories.
- **observability**: Initializes OpenTelemetry tracing and manages error-only inference dumps.
- **pruneagent**: Investigates local-only or likely-removable branches via the Service.Investigate method.
- **remotegit**: Provides adapters to interact with remote git services such as GitHub and GitLab.
- **server**: Provides HTTP serving capabilities and multiplexer initialization.
- **syncproj**: Aggregates project selection and discovery information.
- **triage**: Analyzes project logs and job information to produce a summary and root cause analysis.
