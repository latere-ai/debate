// Command codexmock is the mock codex binary used by e2e tests. It
// emits the JSON event stream expected by internal/agent/critic.go's
// CodexCritic.
package main

import (
	"encoding/json"
	"io"
	"os"
)

func main() {
	if err := run(os.Stdout, os.LookupEnv); err != nil {
		os.Exit(1)
	}
}

// run emits the mock codex event stream to w. lookup is os.LookupEnv
// in production; tests inject a fake.
func run(w io.Writer, lookup func(string) (string, bool)) error {
	thread := envOr(lookup, "MOCK_CODEX_THREAD_ID", "mock-thread")
	content := envOr(lookup, "MOCK_CODEX_CONTENT",
		"# Critic 1 - round 1 attacks\n\naspect: security\n")

	if err := emit(w, "thread.started", map[string]any{"thread_id": thread}); err != nil {
		return err
	}
	if err := emit(w, "item.completed", map[string]any{
		"item": map[string]any{
			"id": "1", "type": "agent_message", "content": content,
		},
	}); err != nil {
		return err
	}
	return emit(w, "thread.completed", map[string]any{"thread_id": thread})
}

func envOr(lookup func(string) (string, bool), key, fallback string) string {
	if v, ok := lookup(key); ok && v != "" {
		return v
	}
	return fallback
}

func emit(w io.Writer, typ string, extra map[string]any) error {
	out := map[string]any{"type": typ}
	for k, v := range extra {
		out[k] = v
	}
	b, err := json.Marshal(out)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}
