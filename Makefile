SHELL := /bin/bash
BIN   := bin/debate
PKG   := ./cmd/debate

.PHONY: all build install test lint vet clean

all: lint vet test build

build:
	mkdir -p bin
	go build -o $(BIN) $(PKG)

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
