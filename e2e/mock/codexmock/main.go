// Command codexmock is the mock codex binary used by e2e tests. It
// emits the JSON event stream expected by internal/agent/critic.go's
// CodexCritic.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	thread := os.Getenv("MOCK_CODEX_THREAD_ID")
	if thread == "" {
		thread = "mock-thread"
	}
	content := os.Getenv("MOCK_CODEX_CONTENT")
	if content == "" {
		content = "# Critic 1 — round 1 attacks\n\naspect: security\n"
	}

	emit("thread.started", map[string]any{"thread_id": thread})
	emit("item.completed", map[string]any{
		"item": map[string]any{
			"id": "1", "type": "agent_message", "content": content,
		},
	})
	emit("thread.completed", map[string]any{"thread_id": thread})
}

func emit(typ string, extra map[string]any) {
	out := map[string]any{"type": typ}
	for k, v := range extra {
		out[k] = v
	}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}
