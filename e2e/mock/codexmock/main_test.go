package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunDefaults(t *testing.T) {
	var buf bytes.Buffer
	lookup := func(string) (string, bool) { return "", false }
	if err := run(&buf, lookup); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		`"thread.started"`,
		`"thread_id":"mock-thread"`,
		`"item.completed"`,
		`"agent_message"`,
		`"thread.completed"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestRunCustomEnv(t *testing.T) {
	var buf bytes.Buffer
	lookup := func(k string) (string, bool) {
		switch k {
		case "MOCK_CODEX_THREAD_ID":
			return "tid-42", true
		case "MOCK_CODEX_CONTENT":
			return "custom critic content", true
		}
		return "", false
	}
	if err := run(&buf, lookup); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"tid-42"`) {
		t.Errorf("custom thread id missing: %s", out)
	}
	if !strings.Contains(out, "custom critic content") {
		t.Errorf("custom content missing: %s", out)
	}
}

func TestEnvOrFallback(t *testing.T) {
	lookup := func(k string) (string, bool) {
		if k == "set-empty" {
			return "", true
		}
		if k == "set-value" {
			return "value", true
		}
		return "", false
	}
	if got := envOr(lookup, "missing", "fallback"); got != "fallback" {
		t.Errorf("missing key: got %q", got)
	}
	if got := envOr(lookup, "set-empty", "fallback"); got != "fallback" {
		t.Errorf("empty value: got %q", got)
	}
	if got := envOr(lookup, "set-value", "fallback"); got != "value" {
		t.Errorf("set value: got %q", got)
	}
}
