package agent

import (
	"strings"
	"testing"
)

// TestFormatClaudeStreamEventToolUse pins the operator-facing string
// for a claude tool_use event. Exact format is the contract a user
// reads in --log-mode verbose, so a regression here is a UX
// regression.
func TestFormatClaudeStreamEventToolUse(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
	}{
		{
			"Read file_path",
			`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"x","name":"Read","input":{"file_path":"/tmp/x.go"}}]}}`,
			"  → Read: /tmp/x.go",
		},
		{
			"Bash command",
			`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"x","name":"Bash","input":{"command":"git status"}}]}}`,
			"  → Bash: git status",
		},
		{
			"Grep pattern",
			`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"x","name":"Grep","input":{"pattern":"TODO"}}]}}`,
			"  → Grep: TODO",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatClaudeStreamEvent([]byte(tc.line)); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatClaudeStreamEventThinkingAndText(t *testing.T) {
	thinking := `{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"First I will look at the diff\nThen check the README"}]}}`
	if got := FormatClaudeStreamEvent([]byte(thinking)); !strings.HasPrefix(got, "  thinking: ") {
		t.Errorf("thinking should be prefixed: %q", got)
	}
	if got := FormatClaudeStreamEvent([]byte(thinking)); !strings.Contains(got, "First I will look") {
		t.Errorf("thinking content missing: %q", got)
	}

	text := `{"type":"assistant","message":{"content":[{"type":"text","text":"Here is my reply\nsecond line"}]}}`
	if got := FormatClaudeStreamEvent([]byte(text)); !strings.HasPrefix(got, "  text: ") {
		t.Errorf("text should be prefixed: %q", got)
	}
}

// TestFormatClaudeStreamEventDrops covers the silence contract: a
// stream-json event the user does not need to see (system init,
// user/tool_result, the final result event) must NOT produce a
// progress line. Without this, verbose output is dominated by
// noise.
func TestFormatClaudeStreamEventDrops(t *testing.T) {
	for _, line := range []string{
		`{"type":"system","subtype":"init","session_id":"x"}`,
		`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"x","content":"output"}]}}`,
		`{"type":"result","subtype":"success","result":"done","is_error":false}`,
		`not even json`,
	} {
		if got := FormatClaudeStreamEvent([]byte(line)); got != "" {
			t.Errorf("expected drop for %q, got %q", line, got)
		}
	}
}

func TestFormatClaudeStreamEventLongInputClipped(t *testing.T) {
	long := strings.Repeat("x", 200)
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"` + long + `"}}]}}`
	got := FormatClaudeStreamEvent([]byte(line))
	if len(got) > 100 {
		t.Errorf("expected clip; got %d chars: %q", len(got), got)
	}
	if !strings.Contains(got, "…") {
		t.Errorf("expected ellipsis on clipped output: %q", got)
	}
}

func TestFormatCodexStreamEventToolCall(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
	}{
		{
			"function_call with input.path",
			`{"type":"item.completed","item":{"type":"function_call","name":"read_file","input":{"path":"/etc/hosts"}}}`,
			"  → read_file: /etc/hosts",
		},
		{
			"command_execution",
			`{"type":"item.completed","item":{"type":"command_execution","command":"ls -la"}}`,
			"  → shell: ls -la",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatCodexStreamEvent([]byte(tc.line)); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatCodexStreamEventDrops(t *testing.T) {
	for _, line := range []string{
		`{"type":"thread.started","thread_id":"x"}`,
		`{"type":"turn.completed","usage":{"input_tokens":1}}`,
		`{"type":"item.completed","item":{"type":"agent_message","text":"hi"}}`,
	} {
		if got := FormatCodexStreamEvent([]byte(line)); got != "" {
			t.Errorf("expected drop for %q, got %q", line, got)
		}
	}
}

// TestDecodeClaudeStreamResult verifies the result line is
// extracted from a multi-line stream-json blob, and that an
// intermediate result-shaped line is overridden by the final one
// (defensive against bizarre CLI output that emits multiple).
func TestDecodeClaudeStreamResult(t *testing.T) {
	stdout := []byte(`{"type":"system","subtype":"init"}
{"type":"assistant","message":{"content":[{"type":"text","text":"reply"}]}}
{"type":"result","subtype":"success","session_id":"sess-1","result":"final body","is_error":false,"total_cost_usd":0.04,"usage":{"input_tokens":100,"output_tokens":20}}
`)
	got, err := decodeClaudeStreamResult(stdout)
	if err != nil {
		t.Fatal(err)
	}
	if got.SessionID != "sess-1" {
		t.Errorf("SessionID: got %q, want sess-1", got.SessionID)
	}
	if got.Result != "final body" {
		t.Errorf("Result: got %q", got.Result)
	}
	if got.TotalCostUSD != 0.04 {
		t.Errorf("TotalCostUSD: got %v", got.TotalCostUSD)
	}
	if got.Usage.InputTokens != 100 {
		t.Errorf("Usage.InputTokens: got %d", got.Usage.InputTokens)
	}
}

func TestDecodeClaudeStreamResultNoneFound(t *testing.T) {
	stdout := []byte(`{"type":"system","subtype":"init"}
{"type":"assistant","message":{"content":[{"type":"text","text":"reply"}]}}
`)
	if _, err := decodeClaudeStreamResult(stdout); err == nil {
		t.Error("expected error when no result event present")
	}
}
