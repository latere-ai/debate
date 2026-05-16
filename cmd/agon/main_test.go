package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"latere.ai/x/agon/internal/cli"
)

func TestExitCodeFor(t *testing.T) {
	if got := exitCodeFor(nil); got != 1 {
		t.Errorf("nil err: got %d, want 1", got)
	}
	if got := exitCodeFor(errors.New("boom")); got != 1 {
		t.Errorf("plain err: got %d", got)
	}
	pe := &cli.PreflightError{Code: 42}
	if got := exitCodeFor(pe); got != 42 {
		t.Errorf("preflight: got %d", got)
	}
	// Wrapped preflight error still surfaces the code.
	wrapped := fmt.Errorf("wrapped: %w", &cli.PreflightError{Code: 7})
	if got := exitCodeFor(wrapped); got != 7 {
		t.Errorf("wrapped preflight: got %d, want 7", got)
	}
}

func TestTaskSource(t *testing.T) {
	cases := []struct {
		name string
		in   *cli.Flags
		want string
	}{
		{"flag", &cli.Flags{TaskContext: "x"}, "flag"},
		{"transcript", &cli.Flags{Transcript: "/p"}, "transcript"},
		{"session-id", &cli.Flags{SessionID: "abc"}, "session-id-resume"},
		{"unknown", &cli.Flags{}, "unknown"},
		// Precedence: TaskContext wins over Transcript / SessionID.
		{"flag-wins", &cli.Flags{TaskContext: "x", Transcript: "/p", SessionID: "y"}, "flag"},
		// Transcript wins over SessionID.
		{"transcript-wins", &cli.Flags{Transcript: "/p", SessionID: "y"}, "transcript"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := taskSource(c.in); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestRealMain_Version(t *testing.T) {
	var buf strings.Builder
	code := realMain([]string{"--version"}, &buf, &buf)
	if code != 0 {
		t.Errorf("--version exit code: got %d, want 0", code)
	}
	if !strings.Contains(buf.String(), "agon") {
		t.Errorf("--version output should mention agon; got %q", buf.String())
	}
}

func TestRealMain_Help(t *testing.T) {
	var buf strings.Builder
	code := realMain([]string{"--help"}, &buf, &buf)
	if code != 0 {
		t.Errorf("--help exit code: got %d, want 0", code)
	}
}

func TestRealMain_BareShowsHelp(t *testing.T) {
	// Bare invocation with no env triggers the help fast-path.
	t.Setenv("AGON_TASK_CONTEXT", "")
	t.Setenv("AGON_SESSION_ID", "")
	t.Setenv("AGON_TRANSCRIPT", "")

	var buf strings.Builder
	code := realMain(nil, &buf, &buf)
	if code != 0 {
		t.Errorf("bare exit code: got %d, want 0", code)
	}
}

func TestRealMain_PreflightExitCode(t *testing.T) {
	// --judge llm is rejected by preflight (v0 only supports 'none').
	var buf strings.Builder
	code := realMain([]string{
		"--task-context", "x",
		"--judge", "llm",
	}, &buf, &buf)
	if code == 0 {
		t.Errorf("expected non-zero exit for --judge llm; got %d", code)
	}
	if !strings.Contains(buf.String(), "agon:") {
		t.Errorf("error line should be prefixed with 'agon:'; got %q", buf.String())
	}
}
