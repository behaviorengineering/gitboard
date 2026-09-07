# Operator Briefing: Architecture Boundary Decisions

Hello, Operator. The Typology refine has identified significant "drift" between our intended architecture and the actual Go import graph. We have 11 specific boundary violations that the automation is not allowed to "fix" by simply adding them to the catalog. These require your leadership.

### The Situation
The current catalog (the intended design) is cleaner than the code. The code is making cross-slice imports that we haven't officially sanctioned via `sliceBindings`. If we just add these bindings to the catalog, we are "lying" to ourselves about the complexity of the system.

### Your Decision Points
You must choose one of three paths for each finding:

1.  **Approve the Binding**: If the coupling is intentional and necessary for the feature to work, update the `refined_catalog_yaml` to include the `sliceBinding`. This formalizes the dependency.
2.  **Refactor the Code**: If the coupling is accidental or violates our domain boundaries (e.g., a UI component reaching too deep into a domain), instruct the developers to refactor the code to remove the import.
3.  **Accept as Debt**: If the coupling is a temporary necessity for a deadline, leave it as is and ensure it is tracked in the `journey_md` debt table.

### Priority Areas
- **The Dashboard (`internal/dashboard`)**: It is currently reaching into `config`, `gitboard` (via `board`), and `remotegit`. Is the dashboard a "god object" that needs these, or should it use interfaces?
- **The CLI Entry Point (`cmd/gitboard`)**: It is pulling in `localgit` and `remotegit` directly. This is common for a CLI runner, but we must decide if this should be a formal binding.
- **Triage Automation (`internal/pruneagent`)**: This agent is reaching into `localgit` and `remotegit`. This is likely the core of the automation, but we need to formalize these links.

Check the `journey_md` technical debt table for the full list of evidence.