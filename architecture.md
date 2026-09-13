<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: mission.md](mission.md) · [Next: conventions.md](conventions.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

markdown: |
# Architecture
> Teaching story. Living project architecture for humans and review grounding. Not Typology seed evidence (see `evidence/typology/architecture_brief.md`).
The system is organized into distinct functional slices that define its bounded contexts and operational responsibilities. This architecture prioritizes clear ownership and separation of concerns through a structured hierarchy of domain-driven slices and supporting libraries.
### System Structure
The high-level shape of the system is defined by several core slices:
Board Slice: Acts as the foundational domain layer, providing shared JSON data types used across the dashboard UI, forge adapters, and server handlers.
CLI Execution Slice: Manages the execution of external processes, providing a controlled interface for system-level operations via `os/exec`.
Dashboard Slice: Serves as the aggregation layer, synthesizing various views and investigation results into a cohesive interface for system oversight.
This modular approach allows for independent evolution of components while maintaining strict contracts between slices, ensuring that the system remains extensible and maintainable as new capabilities are added.
