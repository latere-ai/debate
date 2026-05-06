SHELL := /bin/bash
BIN   := bin/debate
PKG   := ./cmd/debate

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: all pre lint vet test build install clean probe release-check

all: pre test build

# pre — golangci-lint v2 vet runs before build and test.
pre: lint

lint:
	golangci-lint run ./...

vet:
	go vet ./...

test: pre
	go test -race -timeout 120s ./...

build: pre
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

install:
	go install $(PKG)

clean:
	rm -rf bin coverage.txt

probe:
	@for s in scripts/probes/*.sh; do \
	  printf '== %s ==\n' "$$s"; \
	  "$$s" || exit $$?; \
	done

# release-check is the local pre-tag gate: lint, vet, test, build,
# version smoke. CI runs the same.
release-check: pre vet test build
	@./$(BIN) --version
	@echo "release-check: OK"
