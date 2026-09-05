# Cluster Proposal: gitboard

  ## Proposed merges
  - **Merge `observability` into `gitboard`**: `observability` is a sole importer of nothing but is imported only by `gitboard`.
  - **Merge `server` into `gitboard`**: `server` is a sole importer (only called by `gitboard`).
  - **Merge `syncproj` into `gitboard`**: `syncproj` is a sole importer (only called by `gitboard`).
  - **Merge `cliexec` into `gitboard`**: `cliexec` is a CLI surface/utility used by multiple domain packages; it should be consolidated into the primary entry point or the domain it serves. Given its high in-degree from domain packages, it functions as a shared execution surface for the `gitboard` application.

  ## Proposed renames
  - `internal/localgit` $\rightarrow$ `internal/git/local` (Align with domain-driven naming)
  - `internal/remotegit` $\rightarrow$ `internal/git/remote` (Align with domain-driven naming)
  - `internal/llm` $\rightarrow$ `internal/ai` (Capability renaming to avoid "LLM" as a pillar)

  ## Anti-pattern findings
  - **Capability as Pillar**: `llm` is currently treated as a standalone domain slice. It should be a capability within a broader domain (e.g., `triage` or `board`).
  - **CLI as Peer Slice**: `cliexec` is currently a peer slice. CLIs are surfaces, not bounded contexts.
  - **Temporal/Functional Slices**: `pruneagent` and `triage` appear to be functional/agentic roles. If they are purely temporal stages of a pipeline, they should be merged into the domain they operate on (e.g., `board` or `git`).

  ## Boundary debt
  | Type | Description |
  |------|-------------|
  | Structural | `cliexec` and `gitboard` (cmd) are split, but `cliexec` is a dependency for many, suggesting it's a shared utility being treated as a domain. |
  | Domain | `llm` is a capability being promoted to a domain pillar. |
  | Domain | `pruneagent` and `triage` lack clear domain ownership and appear to be task-based slices. |

  ## Rationale
  The current topology is "flat," treating every functional component (LLM, Triage, PruneAgent) as a peer bounded context. This leads to high coupling in the `gitboard` command and `config` package. By merging sole-importer packages into the main application (`gitboard`) and reclassifying capabilities (`llm`) and surfaces (`cliexec`) as components of existing domains, we reduce the number of slices and align the architecture with DDD principles.