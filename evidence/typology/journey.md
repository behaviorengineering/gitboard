# Journey: Typology Refinement (gitboard)

## Status
Refinement complete. Applied cluster proposal to consolidate the topology and reduce slice sprawl.

## Decisions Taken
1. **Consolidation**: Merged `observability`, `server`, `syncproj`, and `cliexec` into the `gitboard` slice to reflect their role as application-level surfaces/utilities rather than independent bounded contexts.
2. **Renaming**: 
   - `internal/localgit` $\rightarrow$ `internal/git/local`
   - `internal/remotegit` $\rightarrow$ `internal/git/remote`
   - `internal/llm` $\rightarrow$ `internal/ai`
3. **Capability Reclassification**: Moved `llm` (now `ai`), `triage`, and `pruneagent` into the `gitboard` domain scope to prevent "Capability as Pillar" anti-pattern.
4. **Surface Alignment**: Ensured all `cmd/` and `internal/cliexec` paths are categorized under `surfaces`.

## Technical Debt and Boundary Violations

| Type | Description |
|------|-------------|
| Structural | `cliexec` is still a shared utility dependency for multiple domain packages but is now housed within the `gitboard` slice. |
| Domain | `triage` and `pruneagent` are currently grouped under `gitboard` but represent functional agentic roles that may eventually require their own domain boundaries. |
| Domain | `ai` is a capability being treated as a component within the main domain rather than a standalone service. |