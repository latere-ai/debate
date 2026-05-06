SHELL := /bin/bash
BIN   := bin/debate
PKG   := ./cmd/debate

.PHONY: all build install test lint vet clean probe

all: lint vet test build

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

install:
	go install $(PKG)

test:
	go test ./...

lint:
	golangci-lint run ./...

vet:
	go vet ./...

clean:
	rm -rf bin coverage.txt

probe:
	@for s in scripts/probes/*.sh; do \
	  printf '== %s ==\n' "$$s"; \
	  "$$s" || exit $$?; \
	done
