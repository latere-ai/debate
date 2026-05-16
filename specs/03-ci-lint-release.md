# Spec 03 - CI, lint, version, release

> **Status: ✅ implemented.**
> Implementation spec for `agon`. See [01-overview.md](01-overview.md) for design intent.

**Depends on:** [02](02-go-module.md).
**Consumed by:** [27](27-release.md).

## Scope

In: GitHub Actions workflow for build/test/lint on push and PR; `golangci-lint` config; `--version` ldflags wiring; `goreleaser` config for tagged releases.

Out: actual release execution (that's [27](27-release.md)); CLI surface ([04](04-cli-flags.md)).

## CI matrix

`.github/workflows/ci.yml` runs on push and pull_request to `main`:

| job | os | go | steps |
|---|---|---|---|
| `build-test` | `ubuntu-latest`, `macos-latest` | `1.26.x` | `make all`, upload `bin/agon` artifact |
| `lint` | `ubuntu-latest` | `1.26.x` | `make lint` |

Cache: `~/go/pkg/mod` keyed by `go.sum`.

CI fails if any of: `go vet`, `golangci-lint run`, `go test ./...`, or `go build ./cmd/agon` fails.

## golangci-lint config

`.golangci.yml`:

```yaml
run:
  timeout: 3m
  go: "1.26"
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - revive
    - gofumpt
    - gocritic
    - misspell
    - prealloc
    - unconvert
issues:
  exclude-dirs:
    - testdata
```

No bespoke linters in v0; the defaults plus `revive`/`gocritic` catch most things.

## --version wiring

`cmd/agon/main.go` declares:

```go
package main

var (
    version = "dev"     // set via -ldflags
    commit  = "none"
    date    = "unknown"
)
```

Build invocation:

```
go build -ldflags "\
  -X main.version=$(git describe --tags --always --dirty) \
  -X main.commit=$(git rev-parse --short HEAD) \
  -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/agon ./cmd/agon
```

`agon --version` prints `agon <version> (<commit>, <date>)` and exits 0. The flag is wired in [04](04-cli-flags.md) and consumes these vars.

## Release with goreleaser

`.goreleaser.yaml`:

```yaml
version: 2
project_name: agon
before:
  hooks:
    - go mod tidy
builds:
  - id: agon
    main: ./cmd/agon
    binary: agon
    env: [CGO_ENABLED=0]
    goos: [linux, darwin]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
archives:
  - format: tar.gz
    name_template: "debate_{{.Version}}_{{.Os}}_{{.Arch}}"
    files:
      - LICENSE
      - README.md
checksum:
  name_template: "checksums.txt"
release:
  draft: true
  github:
    owner: latere-ai
    name: agon
```

Tags: `v0.0.1` style. Tag pushes trigger a separate `release.yml` GitHub Action that runs `goreleaser release --clean`.

## Test contract

- CI green on a fresh clone with the [02](02-go-module.md) skeleton in place.
- `goreleaser check` passes locally.
- `bin/agon --version` prints a non-`dev` string when built via the `make build`/`goreleaser` path.

## Acceptance criteria

- [x] `.github/workflows/ci.yml` exists; passes against the [02](02-go-module.md) baseline.
- [x] `.golangci.yml` exists and `make lint` is clean.
- [x] `cmd/agon/main.go` exposes `version`, `commit`, `date` as set-via-ldflags vars.
- [x] `.goreleaser.yaml` validates with `goreleaser check`.
- [x] Tagging `v0.0.1-rc1` produces a draft release with darwin/linux × amd64/arm64 tarballs.
