# Weaknesses

- **Capability Leak**: The `llm` slice functions as a top-level slice despite being a capability used by other domains.
- **Fragmented Logic**: CLI execution logic (`cliexec`) is currently distributed; it requires future migration into domain-specific adapters.
- **Documentation Gaps**: Several packages lack comprehensive overview documentation and component-specific markdown files.