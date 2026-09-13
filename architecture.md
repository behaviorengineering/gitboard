<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: mission.md](mission.md) · [Next: conventions.md](conventions.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

yaml
markdown: |
# Architecture
> Teaching story. Living project architecture for humans and review grounding. Not Typology seed evidence (see `evidence/typology/architecture_brief.md`).
The Gitboard system is organized as a collection of specialized Go packages that coordinate to bridge local git state with remote forge visibility. Rather than a monolithic engine, the architecture functions through a series of discrete roles:
### Core Coordination and State
The system centers on the interaction between syncproj (managing remote git synchronization) and the server, which provides the backbone for data access. The localgit and remotegit packages act as the primary interfaces for state, ensuring that the local environment and remote repositories are accurately represented.
### Intelligence and Triage
A key layer of the architecture involves the triage and llm packages. These components work in tandem to process incoming data—such as failed jobs or open MRs—and provide intelligent insights. The pruneagent assists in maintaining the hygiene of the local state by managing lifecycle tasks.
### User Interface and Observability
The dashboard provides the human-facing view of the system, consuming data from various internal sources to present a unified board. This is supported by observability tools that ensure the system's internal health is transparent.
### Package Topology
The system is composed of 13 distinct packages, including `cmd/gitboard`, `internal/board`, `internal/cliexec`, `internal/config`, `internal/dashboard`, `internal/llm`, `internal/localgit`, `internal/observability`, `internal/pruneagent`, `internal/remotegit`, `internal/server`, `internal/syncproj`, and `internal/triage`. This modularity allows for independent evolution of the sync logic, the intelligence layer, and the presentation layer.
