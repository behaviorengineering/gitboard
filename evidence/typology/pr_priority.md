# Architecture Alignment: Formalizing Boundary Dependencies

## Overview
This PR addresses the divergence between our intended architecture and the actual Go import graph in the `gitboard` repository. While the recent refinement consolidated several packages into cleaner domains (like `triage` and `gitboard`), the underlying code still contains several cross-slice imports that are not yet formally recognized in our Typology catalog.

## Why This Matters
In a bounded-context architecture, we want to be very intentional about how slices talk to one another. Currently, we have 11 "hidden" connections. If we don't decide whether these are intentional design choices or technical debt, our architecture map becomes an unreliable source of truth.

## The Human Choice
We are presenting these findings not as errors to be automatically fixed, but as **architectural decisions** for the team to make. We need to decide:
1. Do we **formalize** these connections by adding `sliceBindings` (approved allowed couplings between two bounded contexts)?
2. Do we **refactor** the code to remove these imports?
3. Do we **accept** them as temporary technical debt?

## Key Areas of Impact
* **UI Layer (`internal/dashboard`)**: Currently reaching into `config`, `gitboard` (domain), and `remotegit`.
* **CLI Runner (`cmd/gitboard`)**: Directly interacting with `localgit` and `remotegit`.
* **Automation Agent (`internal/pruneagent`)**: Reaching into `localgit` and `remotegit` to perform triage tasks.
* **Remote Logic (`internal/remotegit`)**: Reaching back into `config` and `gitboard` (domain).

Please review the `journey_md` for the full list of violations and provide direction on which bindings to approve and which code to refactor.