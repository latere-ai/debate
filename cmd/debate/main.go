// Command debate is the adversarial review CLI for Claude Code coding
// sessions. The full design lives in specs/01-overview.md.
package main

import (
	"fmt"
	"os"
)

// Set via -ldflags by goreleaser / Makefile.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("debate %s (%s, %s)\n", version, commit, date)
		os.Exit(0)
	}
	// Real CLI wiring lands in spec 04.
	os.Exit(0)
}
