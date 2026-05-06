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
	long := strings.Repeat("x", 300)
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"` + long + `"}}]}}`
	got := FormatClaudeStreamEvent([]byte(line))
	if !strings.Contains(got, "…") {
		t.Errorf("expected ellipsis on clipped output: %q", got)
	}
	// Width budget is summaryWidth+prefix; assert the content body
	// stays under that ceiling (allow some slack for the prefix).
	if len(got) > summaryWidth+30 {
		t.Errorf("clip too wide; got %d chars: %q", len(got), got)
	}
}

// TestFormatClaudeStreamEventEmptyThinkingDropped: a thinking block
// with empty content (claude code emits these as block-start
// markers / partials) must NOT produce a "thinking:" line.
// Reproduces the noisy "thinking:" with-no-body output a real run
// hit.
func TestFormatClaudeStreamEventEmptyThinkingDropped(t *testing.T) {
	for _, line := range []string{
		`{"type":"assistant","message":{"content":[{"type":"thinking","thinking":""}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"thinking","thinking":"   \n  "}]}}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":""}]}}`,
	} {
		if got := FormatClaudeStreamEvent([]byte(line)); got != "" {
			t.Errorf("empty content should drop; got %q", got)
		}
	}
}

// TestFormatClaudeStreamEventLongPathFitsInBudget asserts a typical
// absolute path stays intact: was clipped at 80 chars before, now
// fits because summaryWidth is 120.
func TestFormatClaudeStreamEventLongPathFitsInBudget(t *testing.T) {
	path := "/Users/changkun/dev/changkun.de/agents-byzantine-tolerance/results/07_debate/README.md"
	line := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"` + path + `"}}]}}`
	got := FormatClaudeStreamEvent([]byte(line))
	if strings.Contains(got, "…") {
		t.Errorf("path of length %d should fit in budget %d, got clipped: %q", len(path), summaryWidth, got)
	}
	if !strings.Contains(got, path) {
		t.Errorf("path missing from output: %q", got)
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
		// item.started fires when codex BEGINS a command - this is
		// the live signal that was missing during long critic calls.
		// Captured from real codex-cli 0.128.0 stream.
		{
			"item.started command_execution",
			`{"type":"item.started","item":{"id":"item_1","type":"command_execution","command":"/bin/zsh -lc 'wc -l /etc/hosts'","status":"in_progress"}}`,
			"  → shell: /bin/zsh -lc 'wc -l /etc/hosts'",
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

// TestFormatCodexStreamEventReasoning surfaces the agent's
// reasoning summaries as "thinking:" lines. Without this, codex
// critic calls using o1/o3-style reasoning models showed nothing
// for the entire 100+ second wait.
func TestFormatCodexStreamEventReasoning(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
	}{
		{
			"reasoning with summary",
			`{"type":"item.completed","item":{"type":"reasoning","summary":"Plan: read diff, scan README for inconsistencies"}}`,
			"  thinking: Plan: read diff, scan README for inconsistencies",
		},
		{
			"agent_reasoning text field",
			`{"type":"item.completed","item":{"type":"agent_reasoning","text":"Looking at experiments/07_debate.py first"}}`,
			"  thinking: Looking at experiments/07_debate.py first",
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
