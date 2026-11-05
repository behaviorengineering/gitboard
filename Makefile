.DEFAULT_GOAL := help

.PHONY: help

help: ## Point at the Taskfile runner
	@printf '  Gitboard verbs live in Taskfile.yml. Run:\n\n'
	@printf '    go tool task          # list local-dev tasks\n'
	@printf '    go tool task ci       # local quality gate\n'
	@printf '    go tool task serve    # process-compose TUI on :1325\n'
