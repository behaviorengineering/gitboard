.PHONY: help build test vet ci init sync serve serve-down conformity bench-net

BINARY := bin/gitboard
CONFIG ?= $(HOME)/.config/gitboard/config.yaml
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

help:
	@echo "gitboard - GitLab + GitHub project dashboard (gh + glab)"
	@echo ""
	@echo "  make build       Build $(BINARY)"
	@echo "  make test        go test ./..."
	@echo "  make vet         go vet ./..."
	@echo "  make ci          tidy + gofmt + vet + race tests + build"
	@echo "  make conformity  Network retrieval contract tests across forges"
	@echo "  make bench-net   Network call-count / latency microbenchmarks"
	@echo "  make init        Create $(CONFIG) if missing"
	@echo "  make sync        Discover repos and select tracked projects"
	@echo "  make serve       process-compose TUI (:1325); creates config if missing; rebuilds on file changes"
	@echo "  make serve-down  Stop this Gitboard process-compose project"

build:
	@mkdir -p $(dir $(BINARY))
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/gitboard

test:
	go test ./...

vet:
	go vet ./...

conformity:
	go test ./internal/conformity/ -count=1 -v

bench-net:
	go test ./internal/benchnet/ -bench=. -benchmem -count=1

ci:
	@cp go.mod go.mod.bak && cp go.sum go.sum.bak
	go mod tidy
	@diff -u go.mod.bak go.mod && diff -u go.sum.bak go.sum
	@rm -f go.mod.bak go.sum.bak
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:" && gofmt -l . && exit 1)
	go vet ./...
	go test -race -count=1 ./...
	go build ./...

init: build
	./$(BINARY) init -config $(CONFIG)

sync: build
	./$(BINARY) sync -config $(CONFIG)

serve: build
	@test -f $(CONFIG) || ./$(BINARY) init -config $(CONFIG)
	chmod +x scripts/pc-up.sh scripts/pc-down.sh
	./scripts/pc-up.sh

serve-down:
	chmod +x scripts/pc-down.sh
	./scripts/pc-down.sh
