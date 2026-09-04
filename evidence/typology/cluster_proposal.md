# Cluster Proposal: gitboard

  ## Proposed merges
  - **gitboard-core**: Merge `server`, `syncproj`, and `observability` into the `gitboard` domain. These are currently sole-importer leaves or tightly coupled to the main entry point.
  - **git-engine**: Merge `localgit` and `remotegit` into a single git-handling domain. They share a job family and high mutual/caller coupling.
  - **gitboard-cli**: Merge `cliexec` into `gitboard` as a surface/utility rather than a peer slice.

  ## Proposed renames
  - `llm` $\rightarrow$ `llm-gateway` (to signal it is a capability, not a domain pillar).
  - `pruneagent` $\rightarrow$ `gitboard-pruner` (to align with the domain owner).
  - `triage` $\rightarrow$ `gitboard-triage` (to align with the domain owner).

  ## Anti-pattern findings
  - **Capability as Pillar**: `llm` is currently a peer slice; it should be a capability used by domain slices (like `triage` or `pruneagent`).
  - **CLI as Peer Slice**: `cliexec` is treated as a top-level domain slice; it is a surface/utility and should be nested under the domain it serves.
  - **Temporal/Task Slices**: `pruneagent` and `triage` are currently peer slices; they represent specific tasks/stages and should be components within a broader domain context.

  ## Boundary debt
  | Violation | Description |
  | :--- | :--- |
  | Capability Leak | `llm` is exposed as a primary domain slice instead of a supporting service. |
  | Fragmented Git Logic | `localgit` and `remotegit` are split, creating unnecessary boundary crossings for git operations. |
  | Surface/Domain Blur | `cliexec` sits at the same hierarchy level as core domain logic. |

  ## Rationale
  The current topology is "flat," treating every functional package as a peer bounded context. This leads to a high-degree hub (`gitboard`) that must import almost every other slice. By consolidating git-specific logic, moving capabilities (`llm`) to a supporting role, and collapsing task-specific agents (`pruneagent`) into the domain, we reduce the cognitive load and the number of explicit boundaries that must be managed.