package main

import (
	"errors"
	"testing"

	"latere.ai/x/debate/internal/cli"
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
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := taskSource(c.in); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
