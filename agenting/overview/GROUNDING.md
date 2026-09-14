markdown: |
# Grounding
The Majordomo bootstrap establishes the foundational mission and architectural constraints for all autonomous agents within the Gitboard ecosystem. This grounding ensures that agentic operations remain aligned with the core objective: providing a high-fidelity, local-first interface for managing forge-based development workflows (GitLab/GitHub).
## Mission Alignment
Agents operate under the principle of "local-first visibility," transforming raw forge data into actionable intelligence. The mission is to facilitate seamless triage of CI status, merge requests, and worktree states without compromising the developer's local environment or security posture.
## Architectural Typology
The architecture follows the refined Typology catalog, moving beyond raw inventory toward a journey-based orchestration. Agents are categorized by their interaction with the `gitboard` lifecycle:
Discovery Agents: Mapping the current state of the forge and local repository.
Triage Agents: Analyzing failed jobs and open MRs to propose actionable resolutions.
Sync Agents: Maintaining the integrity of the local `gitboard` state against remote forge realities.
All agentic reasoning must preserve `majordomo-reading` markers to ensure traceability across the bootstrap story.