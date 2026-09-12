# Chronology

<!-- majordomo-reading-nav:start -->
**Reading path:** [Prev: weaknesses.md](weaknesses.md) · [Next: README.md](evidence/typology/README.md) · [TOC](README.md)
<!-- majordomo-reading-nav:end -->

Newest first.

- **Typology Refinement (Current)**: 
    - Merged `localgit` and `remotegit` into the `git` slice to reduce adapter sprawl.
    - Merged `pruneagent` and `triage` into the `analysis` slice to consolidate project health/triage logic.
    - Renamed `board` to `dto/board` to explicitly signal its role as a technical data-shape library.
    - Renamed `cliexec` to `runner` to align with its functional role.
    - Identified boundary debt regarding `analysis` imports of `git` and `runner`.
