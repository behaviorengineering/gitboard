# Conventions

## Development Workflow
- **Architecture First:** Changes to package boundaries should be reflected in the Typology catalog before implementation.
- **Slice Integrity:** Maintain clear boundaries between slices. Cross-slice dependencies should be explicit and ideally mediated through defined interfaces or DTOs.
- **Composition Root:** Use `cmd/gitboard` as the primary composition root for wiring dependencies, but aim to keep the entry point logic thin.

## Package Organization
- **Internal Packages:** Domain logic resides in `internal/`.
- **DTOs:** Data transfer objects used for UI communication reside within the `dashboard` slice.
- **Libraries:** Technical utilities (like `config` or `llm`) are treated as shared libraries rather than domain slices.