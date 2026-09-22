.DEFAULT_GOAL := help

.PHONY: help build test vet tidy lint format ci init sync serve serve-down conformity conformity-live bench-net

BINARY := bin/gitboard
CONFIG ?= $(HOME)/.config/gitboard/config.yaml
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

help: ## List available make verbs
	@grep -E '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'

build: ## Build bin/gitboard
	@mkdir -p $(dir $(BINARY))
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/gitboard

test: ## Run unit tests
	go test ./...

vet: ## Run go vet ./...
	go vet ./...

tidy: ## Run go mod tidy
	go mod tidy

format: ## Format Go sources with gofmt
	gofmt -w .

lint: ## Run golangci-lint
	golangci-lint run ./...

conformity: ## Network retrieval contract tests (fake exec)
	go test ./internal/conformity/ -count=1 -v

conformity-live: ## Opt-in live forge Collect (needs credentials)
	GITBOARD_CONFORMITY_LIVE=1 go test ./internal/conformity/ -run 'TestLiveAdapterContract|TestLiveDisabledByDefault' -count=1 -v

bench-net: ## Network call-count / latency microbenchmarks
	go test ./internal/benchnet/ -bench=. -benchmem -count=1

ci: ## tidy + gofmt check + vet + race tests + build
	@cp go.mod go.mod.bak && cp go.sum go.sum.bak
	go mod tidy
	@diff -u go.mod.bak go.mod && diff -u go.sum.bak go.sum
	@rm -f go.mod.bak go.sum.bak
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:" && gofmt -l . && exit 1)
	go vet ./...
	go test -race -count=1 ./...
	go build ./...

init: build ## Create config if missing
	./$(BINARY) init -config $(CONFIG)

sync: build ## Discover repos and select tracked projects
	./$(BINARY) sync -config $(CONFIG)

serve: build ## process-compose TUI (:1325); creates config if missing
	@test -f $(CONFIG) || ./$(BINARY) init -config $(CONFIG)
	chmod +x scripts/pc-up.sh scripts/pc-down.sh
	./scripts/pc-up.sh

serve-down: ## Stop this Gitboard process-compose project
	chmod +x scripts/pc-down.sh
	./scripts/pc-down.sh
