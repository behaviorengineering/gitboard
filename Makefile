.PHONY: help build test vet serve

BINARY := bin/gitboard
PROJECTS ?= projects.yaml

help:
	@echo "gitboard — GitLab + GitHub project dashboard (gh + glab)"
	@echo ""
	@echo "  make build    Build $(BINARY)"
	@echo "  make test     go test ./..."
	@echo "  make vet      go vet ./..."
	@echo "  make serve    Run local dashboard on :1325"

build:
	@mkdir -p $(dir $(BINARY))
	go build -o $(BINARY) ./cmd/gitboard

test:
	go test ./...

vet:
	go vet ./...

serve: build
	@test -f $(PROJECTS) || { echo "copy projects.example.yaml to $(PROJECTS) first" >&2; exit 1; }
	./$(BINARY) -projects $(PROJECTS)
