# Spec 02 - Go module and layout

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) for design intent.

**Depends on:** none.
**Consumed by:** every other implementation spec - this defines the skeleton everything else fills in.

## Scope

In: Go module declaration, repo directory layout, Makefile targets, `.gitignore`, empty-package skeleton.

Out: build/test pipeline ([03](03-ci-lint-release.md)), CLI logic ([04](04-cli-flags.md)), any subprocess driver, the Stop-hook script.

## Module path and toolchain

- Module: `latere.ai/x/debate`.
- Go directive in `go.mod`: `go 1.23` (selected for `slices`, `cmp`, `errors.Join`, integer `for` range).
- No third-party dependencies in v0 except a CLI library and TOML parser; both pinned in [04](04-cli-flags.md) and [05](05-config-file.md). Standard library elsewhere.

## Directory layout

```
debate/
├── cmd/debate/                # main.go: flag wiring only
├── internal/
│   ├── cli/                   # filled by 04, 05, 06
│   ├── input/                 # filled by 07, 08
│   ├── state/                 # filled by 09, 10, 11
│   ├── ledger/                # filled by 12
│   ├── critic/                # filled by 13, 14, 15
│   ├── agent/                 # filled by 16, 17, 18
│   ├── round/                 # filled by 19, 20, 21
│   ├── summary/               # filled by 22, 23
│   └── hook/                  # filled by 24
├── scripts/
│   ├── debate-stop-hook.sh    # filled by 24
│   └── probes/                # filled by 25
├── testdata/                  # fixtures (filled per-spec)
├── specs/                     # this directory
├── Makefile
├── go.mod
├── go.sum
├── .gitignore
├── LICENSE
└── README.md
```

Rules:

- All implementation under `internal/`. No exported library API for v0.
- `cmd/debate/main.go` does flag wiring → `internal/cli.Run(ctx, args)` only. Zero business logic in `cmd`.
- One package per directory; no nested sub-packages inside `internal/<name>` unless a later spec explicitly requires it.

## Makefile

```makefile
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
```

## .gitignore

```
bin/
.debate/
*.test
*.out
coverage.txt
```

## Empty-package skeleton

Each `internal/<pkg>/` ships a single `doc.go`:

```go
// Package <pkg> is filled in by spec NN.
package <pkg>
```

So `go build ./...` succeeds on a fresh clone before any business logic lands.

## Test contract

- `go build ./...` clean.
- `go test ./...` reports zero tests run, no failures.
- `go vet ./...` clean.

## Acceptance criteria

- [x] `go.mod` is `latere.ai/x/debate` at `go 1.23`.
- [x] Every directory in the layout exists, each with at least `doc.go` (where applicable).
- [x] `make build` produces `bin/debate`.
- [x] `bin/debate` exits 0 (placeholder until [04](04-cli-flags.md) replaces `main`).
- [x] `make all` succeeds on a fresh clone with no warnings.
