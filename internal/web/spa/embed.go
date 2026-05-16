// Package spa holds the embedded, Bun-built Agon landing site.
//
// The dist/ directory is a build placeholder in the repository
// (dist/PLACEHOLDER); the Docker build replaces it with the real
// vite-ssg output before `go build`, so `go:embed` picks up real
// assets. `go build ./...` and CI work without a frontend build
// because the placeholder keeps the embed directive satisfiable.
package spa

import "embed"

// FS holds the embedded, Bun-built Agon landing site (vite-ssg dist).
//
//go:embed all:dist
var FS embed.FS
