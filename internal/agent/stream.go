package agent

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatClaudeStreamEvent turns one line of `claude --output-format
// stream-json` output into a one-line human-readable progress
// message, or returns "" when the event has nothing worth
// surfacing. The format is the same shape across critic and
// proposer drivers because claude emits identical event types in
// both --resume and --print modes.
//
// What we surface:
//
//	tool_use   → "  → <Tool>: <input summary>"
//	thinking   → "  thinking: <first line, 80 chars>"
//	text       → "  text: <first line, 80 chars>"
//
// What we drop: system init, user/tool_result events, the final
// result event (the orchestrator already prints its own
// "{role} done in Xs ..." line).
func FormatClaudeStreamEvent(line []byte) string {
	var ev struct {
		Type    string `json:"type"`
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(line, &ev); err != nil {
		return ""
	}
	if ev.Type != "assistant" {
		return ""
	}
	var parts []struct {
		Type     string          `json:"type"`
		Name     string          `json:"name"`
		Input    json.RawMessage `json:"input"`
		Text     string          `json:"text"`
		Thinking string          `json:"thinking"`
	}
	if json.Unmarshal(ev.Message.Content, &parts) != nil {
		return ""
	}
	for _, p := range parts {
		switch p.Type {
		case "tool_use":
			return fmt.Sprintf("  → %s: %s", p.Name, summarizeToolInput(p.Input))
		case "thinking":
			return "  thinking: " + previewLine(p.Thinking)
		case "text":
			return "  text: " + previewLine(p.Text)
		}
	}
	return ""
}

// summarizeToolInput pulls the most operator-useful field out of a
// tool_use input blob. Claude tools expose a small set of common
// keys (file_path / path / command / pattern); we surface the first
// one that's a non-empty string. Falls back to a short JSON
// preview when nothing matches.
func summarizeToolInput(input json.RawMessage) string {
	var generic map[string]any
	if json.Unmarshal(input, &generic) != nil {
		return ""
	}
	for _, key := range []string{"file_path", "path", "command", "pattern", "url", "query"} {
		if v, ok := generic[key].(string); ok && v != "" {
			return clip(v, 80)
		}
	}
	return clip(string(input), 80)
}

// previewLine returns the first line of s, ellipsized at 80 chars.
// Used for thinking/text events where the full body is usually
// multi-paragraph and would dominate the progress stream.
func previewLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return clip(s, 80)
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// FormatCodexStreamEvent is the codex counterpart of
// FormatClaudeStreamEvent. Codex emits item.completed events whose
// item.type discriminates kind: agent_message, tool_call,
// command_execution, etc. We surface tool_call and
// command_execution and drop the rest (agent_message is the final
// answer the orchestrator already records).
func FormatCodexStreamEvent(line []byte) string {
	var ev struct {
		Type string `json:"type"`
		Item struct {
			Type    string          `json:"type"`
			Name    string          `json:"name"`
			Command string          `json:"command"`
			Path    string          `json:"path"`
			Input   json.RawMessage `json:"input"`
			Args    json.RawMessage `json:"args"`
		} `json:"item"`
	}
	if err := json.Unmarshal(line, &ev); err != nil {
		return ""
	}
	if ev.Type != "item.completed" {
		return ""
	}
	switch ev.Item.Type {
	case "tool_call", "function_call":
		summary := ev.Item.Path
		if summary == "" {
			summary = summarizeToolInput(ev.Item.Input)
		}
		if summary == "" {
			summary = summarizeToolInput(ev.Item.Args)
		}
		return fmt.Sprintf("  → %s: %s", firstNonEmpty(ev.Item.Name, "tool"), summary)
	case "command_execution", "shell_command":
		return "  → shell: " + clip(ev.Item.Command, 80)
	}
	return ""
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
