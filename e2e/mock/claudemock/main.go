// Command claudemock is the mock claude binary used by e2e tests.
// It reads a script file from $MOCK_CLAUDE_SCRIPT and emits the JSON
// shape expected by internal/agent/claude_proposer.go and the claude
// critic driver.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type matcher struct {
	ArgsContains string         `json:"args_contains"`
	Stdout       map[string]any `json:"stdout"`
	Stderr       string         `json:"stderr"`
	Exit         int            `json:"exit"`
}

func main() {
	args := strings.Join(os.Args[1:], " ")
	scriptPath, _ := os.LookupEnv("MOCK_CLAUDE_SCRIPT")
	stdout, stderr, code, err := decide(args, scriptPath, os.ReadFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mock-claude:", err)
		os.Exit(2)
	}
	emit(os.Stdout, os.Stderr, stdout, stderr, code)
}

// decide picks the matching response for args from the script at
// scriptPath. read is os.ReadFile in production.
//
// Returns (stdout, stderr, exit, err) where err is non-nil only when
// the script can't be read or parsed (caller exits 2). When no rule
// matches the args, it returns ("", "no rule matched", 3, nil) so the
// caller exits with that code, mirroring the original behaviour.
func decide(args, scriptPath string, read func(string) ([]byte, error)) (
	stdout map[string]any, stderr string, code int, err error,
) {
	if scriptPath == "" {
		// Default: emit a minimal valid claude --output-format json result.
		return map[string]any{
			"type": "result", "subtype": "success",
			"session_id": "mock-session", "result": "ok",
			"is_error": false,
			"usage":    map[string]any{"input_tokens": 1, "output_tokens": 1},
		}, "", 0, nil
	}
	b, err := read(scriptPath)
	if err != nil {
		return nil, "", 0, fmt.Errorf("cannot read script: %w", err)
	}
	var rules []matcher
	if err := json.Unmarshal(b, &rules); err != nil {
		return nil, "", 0, fmt.Errorf("bad script: %w", err)
	}
	for _, m := range rules {
		if m.ArgsContains == "" || strings.Contains(args, m.ArgsContains) {
			return m.Stdout, m.Stderr, m.Exit, nil
		}
	}
	return nil, "no rule matched args: " + args + "\n", 3, nil
}

func emit(out io.Writer, errw io.Writer, stdout map[string]any, stderr string, code int) {
	if stdout != nil {
		b, _ := json.Marshal(stdout)
		_, _ = fmt.Fprintln(out, string(b))
	}
	if stderr != "" {
		_, _ = fmt.Fprint(errw, stderr)
	}
	if code != 0 {
		os.Exit(code)
	}
}
