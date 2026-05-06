// Command debate is the adversarial review CLI for Claude Code coding
// sessions. The full design lives in specs/01-overview.md.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/latere-ai/debate/internal/cli"
	"github.com/latere-ai/debate/internal/hook"
)

// Set via -ldflags by goreleaser / Makefile.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := &cobra.Command{
		Use:           "debate",
		Short:         "Adversarial review for Claude Code coding sessions.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (%s, %s)", version, commit, date),
	}
	root.SetVersionTemplate("debate {{.Version}}\n")

	flags := cli.Bind(root)
	root.RunE = func(_ *cobra.Command, _ []string) error {
		if _, err := cli.Effective(root, flags); err != nil {
			return err
		}
		// Real run lands in spec 19 wired through the orchestrator
		// once cmd/debate is fully integrated; pre-flight is the gate.
		_, err := cli.Preflight(root.Context(), flags)
		if err != nil {
			if errIsRecursion(err) {
				return nil
			}
			return err
		}
		return nil
	}

	root.AddCommand(installHookCmd(), uninstallHookCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "debate:", err)
		os.Exit(exitCodeFor(err))
	}
}

func errIsRecursion(err error) bool {
	return err != nil && err.Error() == "recursion guard triggered"
}

func exitCodeFor(err error) int {
	if pe, ok := err.(*cli.PreflightError); ok && pe != nil {
		return pe.Code
	}
	return 1
}

func installHookCmd() *cobra.Command {
	var scope, scriptPath string
	cmd := &cobra.Command{
		Use:   "install-hook",
		Short: "Install the Stop hook into ~/.claude/settings.json (or project)",
		RunE: func(_ *cobra.Command, _ []string) error {
			s := hook.ScopeUser
			if scope == "project" {
				s = hook.ScopeProject
			}
			if scriptPath == "" {
				scriptPath = hook.LocateScript()
			}
			return hook.Install(s, scriptPath)
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "user", "user | project")
	cmd.Flags().StringVar(&scriptPath, "script-path", "", "explicit path to debate-stop-hook.sh")
	return cmd
}

func uninstallHookCmd() *cobra.Command {
	var scope string
	cmd := &cobra.Command{
		Use:   "uninstall-hook",
		Short: "Remove the Stop hook from ~/.claude/settings.json (or project)",
		RunE: func(_ *cobra.Command, _ []string) error {
			s := hook.ScopeUser
			if scope == "project" {
				s = hook.ScopeProject
			}
			return hook.Uninstall(s)
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "user", "user | project")
	return cmd
}
