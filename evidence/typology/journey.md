# Journey Notes: gitboard Typology Refinement

  ## Status
  Open. Majordomo has reorganized the flat topology into a structured context-centric model, but several boundary violations require human decision or catalog updates.

  ## Decisions Taken

  ### Merge Observability, Server, and Syncproj into Gitboard
  - **Decision**: Folded `observability`, `server`, and `syncproj` into the `gitboard` slice.
  - **Lean**: These packages are specialized logic driven exclusively by the `gitboard` entry point.

  ### Reclassify cliexec as exec-utils
  - **Decision**: Moved `cliexec` from a `kind: cli` surface to an `owns[]` component within a new `exec-utils` slice.
  - **Lean**: `package_contracts` confirms `hasMain: false`, making it a library of runners.

  ### Consolidate Git Logic
  - **Decision**: Grouped `localgit` and `remotegit` under a single `git` slice.
  - **Lean**: Prevents split-brain logic for git operations.

  ## Technical Debt and Boundary Violations

  | Finding | Smell | Alternatives | Lean |
  |---------|-------|--------------|------|
  | 1. `git` -> `board` | Missing SliceBinding for `internal-remotegit` -> `internal-board`. | 1. Add SliceBinding. 2. Decouple via interface. | Add SliceBinding to allow `git` to drive `board` state. |
  | 2. `git` -> `config` | Missing SliceBinding for `internal-remotegit` -> `internal-config`. | 1. Add SliceBinding. 2. Inject config. | Add SliceBinding to allow `git` to read system config. |
  | 3. `gitboard` -> `exec-utils` | Missing SliceBinding for `cmd-gitboard` -> `internal-cliexec`. | 1. Add SliceBinding. 2. Move `cliexec` into `gitboard`. | Add SliceBinding to permit the CLI to use the runner library. |
  | 4. `llm` -> `config` | Missing SliceBinding for `internal-llm` -> `internal-config`. | 1. Add SliceBinding. 2. Inject config. | Add SliceBinding to allow LLM providers to access settings. |
  | **Platform Leakage** | `config` is a high-degree hub used by almost every slice. | 1. Keep as global library. 2. Inject via interfaces. | Keep `config` as a platform utility library for now. |
  | **Orphaned Dashboard** | `dashboard` acts as a middle-man between `server` and `board`. | 1. Keep as service layer. 2. Fold `dashboard` into `server`. | Fold `dashboard` logic into the `server` slice. |