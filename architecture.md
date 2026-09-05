# Architecture

  Gitboard is organized into functional slices that separate domain logic from infrastructure and delivery surfaces.

  ## Bounded Contexts

  | Slice | Objective | Key Components |
  |-------|-----------|----------------|
  | `board` | Provide board functionality | `./internal/board` |
  | `cliexec` | Provide command-line execution capabilities | `./internal/cliexec` (CLI surface) |
  | `config` | Provide configuration management | `./internal/config` |
  | `dashboard` | Provide dashboard functionality | `./internal/dashboard` (UI surface) |
  | `gitboard` | Provide CLI tool and core services | `./cmd/gitboard`, `./internal/observability`, `./internal/server`, `./internal/syncproj` |
  | `llm` | Provide LLM integration capabilities | `./internal/llm` |
  | `git` | Provide git infrastructure services | `./internal/localgit`, `./internal/remotegit` |
  | `pruneagent` | Provide prune agent functionality | `./internal/pruneagent` |
  | `triage` | Provide triage functionality | `./internal/triage` |

  ## System Topology

  The system follows a layered approach where high-level services (like `gitboard` and `dashboard`) consume lower-level infrastructure slices (`config`, `git`, `board`). The `git` slice acts as a critical infrastructure provider for both the dashboard and the prune agent.