# Cluster Proposal: gitboard

  ## Proposed merges
  - **gitboard (main domain)**: Merge `observability`, `server`, and `syncproj` into the `gitboard` context. These are currently sole-importer leaves or tightly coupled to the main entry point.
  - **git-operations (domain slice)**: Group `localgit` and `remotegit` into a single domain slice focused on git primitives.
  - **gitboard-cli (surface)**: Merge `cliexec` and `gitboard` (cmd) into the `gitboard` domain as its primary surface.

  ## Proposed renames
  - `cliexec` $\rightarrow$ `gitboard/surface/cli` (Move from peer slice to domain surface).
  - `remotegit` + `localgit` $\rightarrow$ `git-engine` (Consolidate git primitives).
  - `pruneagent` $\rightarrow$ `gitboard/worker/prune` (Move from peer slice to a domain sub-component/worker).

  ## Anti-pattern findings
  - **Capability as Pillar**: `llm` is currently a peer slice. It is a capability, not a domain. It should be moved under a domain that consumes it (e.g., `triage` or `board`).
  - **Standalone CLI**: `cliexec` is acting as a peer slice. CLIs are surfaces of the domains they invoke, not independent bounded contexts.
  - **Temporal/Functional Slices**: `pruneagent`, `triage`, and `syncproj` are currently peer slices. These appear to be functional tasks or pipeline stages rather than distinct domain models.

  ## Boundary debt
  | Violation | Description |
  | :--- | :--- |
  | Capability Leak | `llm` exists as a top-level domain slice instead of a utility/capability. |
  | Surface Fragmentation | `cliexec` and `gitboard` (cmd) are split, creating redundant CLI entry points. |
  | Task-as-Domain | `pruneagent` and `triage` are defined as domains but represent specific workflows. |

  ## Rationale
  The current topology shows a "hub-and-spoke" model where `gitboard` (cmd) is the center of everything, yet the catalog defines 13 distinct peer slices. This creates massive cognitive load and artificial boundaries. By collapsing the sole-importer leaves (`observability`, `server`, `syncproj`) and reclassifying capabilities (`llm`) and surfaces (`cliexec`) as components of the core `gitboard` domain, we move from a fragmented package list to a true Bounded Context architecture.