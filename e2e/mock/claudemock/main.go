// Command claudemock is the mock claude binary used by e2e tests.
// It reads a script file from $MOCK_CLAUDE_SCRIPT and emits the JSON
// shape expected by internal/agent/claude_proposer.go and the claude
// critic driver.
package main

import (
	"encoding/json"
	"fmt"
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

	script := os.Getenv("MOCK_CLAUDE_SCRIPT")
	if script == "" {
		// Default: emit a minimal valid claude --output-format json result.
		emit(map[string]any{
			"type": "result", "subtype": "success",
			"session_id": "mock-session", "result": "ok",
			"is_error": false,
			"usage": map[string]any{"input_tokens": 1, "output_tokens": 1},
		}, "", 0)
		return
	}
	b, err := os.ReadFile(script)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mock-claude: cannot read script:", err)
		os.Exit(2)
	}
	var rules []matcher
	if err := json.Unmarshal(b, &rules); err != nil {
		fmt.Fprintln(os.Stderr, "mock-claude: bad script:", err)
		os.Exit(2)
	}
	for _, m := range rules {
		if m.ArgsContains == "" || strings.Contains(args, m.ArgsContains) {
			emit(m.Stdout, m.Stderr, m.Exit)
			return
		}
	}
	fmt.Fprintln(os.Stderr, "mock-claude: no rule matched args:", args)
	os.Exit(3)
}

func emit(stdout map[string]any, stderr string, code int) {
	if stdout != nil {
		b, _ := json.Marshal(stdout)
		fmt.Println(string(b))
	}
	if stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}
	os.Exit(code)
}
