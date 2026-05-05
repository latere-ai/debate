// Command debate is the adversarial review CLI for Claude Code coding
// sessions. The full design lives in specs/01-overview.md.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/latere-ai/debate/internal/cli"
)

// Set via -ldflags by goreleaser / Makefile.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd := &cobra.Command{
		Use:           "debate",
		Short:         "Adversarial review for Claude Code coding sessions.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (%s, %s)", version, commit, date),
	}
	cmd.SetVersionTemplate("debate {{.Version}}\n")

	flags := cli.Bind(cmd)
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		cli.ApplyEnv(cmd, flags)
		// Real run lands in spec 06+ (preflight + round loop).
		return nil
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "debate:", err)
		os.Exit(1)
	}
}
