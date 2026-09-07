# Architecture

Gitboard is organized into bounded contexts that separate low-level git operations, configuration, and user interfaces.

## Bounded Contexts

### CLI &amp; Orchestration (`gitboard`)
The primary entry point for repository management and automation. It orchestrates the ecosystem by coordinating configuration, dashboard serving, and automated agents.

### Dashboard &amp; Board (`dashboard`, `board`)
The `dashboard` provides a web-based UI for monitoring data. It relies on the `board` slice to manage the visual representation and state of project boards.

### Git Operations (`git`)
Handles all low-level interactions with local and remote repositories. This context encapsulates both local git state and remote forge data.

### Automation Agents (`pruneagent`, `triage`)
Specialized logic for repository maintenance. `pruneagent` automates branch cleanup, while `triage` assists in decision-making for branch maintenance using LLM capabilities.

### Infrastructure &amp; Support (`config`, `cliexec`)
- `config`: Centralized management of all component settings.
- `cliexec`: An execution adapter for running external processes within the CLI context.

## Boundary Debt
- **Cross-Slice Imports**: Several slices (`dashboard`, `git`, `triage`) currently import `config` without formal `sliceBindings`.
- **Git Abstraction Leaks**: `remotegit` contains unexpected dependencies on `board` and `cliexec`.
- **Implicit Capabilities**: `pruneagent` and `triage` utilize `llm` capabilities via the `gitboard` orchestrator without explicit architectural declarations.