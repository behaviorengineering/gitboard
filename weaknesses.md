# Weaknesses

## Known Architectural Trade-offs
- **Entrypoint Bloat**: The `gitboard` slice currently contains `server` and `observability` components. This is an accepted trade-off to prevent slice sprawl in a small-scale tool.
- **Adapter Classification**: While `localgit` and `remotegit` function as adapters, they currently reside within the `internal/` directory.