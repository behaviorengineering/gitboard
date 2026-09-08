# Proposed merges

### Merge `observability`, `server`, and `syncproj` into `gitboard`
Majordomo proposes consolidating these three packages into the `gitboard` surface. 
- **Why**: Each is a sole importer of the `gitboard` CLI/cmd package (or vice versa in terms of dependency direction), representing specialized logic used exclusively by the main entry point.
- **Rejected**: Keeping them as peer slices creates a fragmented "flat" topology where the primary driver (`gitboard`) is just one of many peers.
- **Cost of alternative**: High cognitive load to navigate multiple small slices that lack independent lifecycle or consumer interest.
- **The lean**: Fold these into the `gitboard` domain or its immediate sub-packages to simplify the bounded context.

### Merge `cliexec` into `gitboard` surfaces
Typology digest proposes moving the `cliexec` logic under the `gitboard` surface.
- **Why**: `cliexec` is an exec-adapter (it has `hasMain: false` and exports `Runner/Exec`). It is heavily consumed by the primary CLI and other domain packages to facilitate command execution.
- **Rejected**: Treating `cliexec` as a standalone "platform" slice.
- **Cost of alternative**: Artificial inflation of the slice count with non-domain utility packages.
- **The lean**: Reclassify `cliexec` as a component within the `gitboard` surface or a supporting library.

# Proposed renames

### Rename `cliexec` to `exec` (within `gitboard` context)
Majordomo proposes renaming the package to reflect its role as a utility for execution rather than a standalone "CLI" entity.

# Anti-pattern findings

### Exec-Adapter as CLI Surface
The current catalog treats `cliexec` as a `kind: cli` surface. 
- **Evidence**: `package_contracts` shows `hasMain: false`. 
- **Violation**: A CLI surface must be a delivery package (a `main` package). `cliexec` is a library of runners.
- **Correction**: Demote `cliexec` from `kind: cli` to `owns[]` or `libraries[]`.

### Capability-based Slicing
`llm` and `triage` are currently treated as peer domain slices.
- **Violation**: These are capabilities/tools rather than domain pillars. 
- **Correction**: These should be viewed as supporting services to the core `board` or `dashboard` domains.

# Boundary debt

| Smell | Alternatives | Lean |
|-------|--------------|------|
| **Fragmented Git Logic**: `localgit` and `remotegit` are split, creating split-brain logic for git operations. | 1. Keep separate (High complexity). 2. Merge into a single `git` domain (Medium cost). | Consolidate into a single `git` domain slice. |
| **Platform Leakage**: `config` is a high-degree hub used by every slice. | 1. Keep as global library. 2. Inject via interfaces. | Keep `config` as a platform utility library. |
| **Orphaned Dashboard**: `dashboard` acts as a middle-man between `server` and `board`. | 1. Keep as service layer. 2. Fold `dashboard` into `server`. | Fold `dashboard` logic into the `server` slice. |

# Rationale

The current repository topology is "flat," treating every package as a peer slice. This obscures the actual hierarchy where `gitboard` (the CLI) and `server` (the API) act as the primary entry points. Majordomo proposes a transition from a package-centric view to a context-centric view. By merging sole-importer utilities and reclassifying exec-adapters, the architecture moves from a collection of 13 disconnected parts toward a structured system of a few core domains (`board`, `git`, `dashboard`) supported by technical libraries (`config`, `observability`).