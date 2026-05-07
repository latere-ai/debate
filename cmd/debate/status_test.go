package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestComputeStatusLine_NoSessionPrintsIdleBar(t *testing.T) {
	// idleBar is a single space, not the empty string: claude's
	// statusLine keeps the previous frame visible across empty
	// outputs. A space forces a render and clears any stale
	// in-flight frame from a previous run.
	if got := computeStatusLine(t.TempDir(), time.Now()); got != idleBar {
		t.Errorf("expected idleBar %q, got %q", idleBar, got)
	}
}

func TestComputeStatusLine_RecentlyFinishedShowsTerminalState(t *testing.T) {
	dir := t.TempDir()
	sess := filepath.Join(dir, ".debate", "sessions", "20260507T120000Z-aaa")
	mustMkdir(t, sess)
	mustWrite(t, filepath.Join(sess, "end.json"),
		`{"termination":{"reason":"steady-state"},`+
			`"stats":{"total_attacks":5,"by_status":{"unresolved":2,"conceded":3}}}`)
	got := computeStatusLine(dir, time.Now())
	for _, want := range []string{"steady-state", "2/5 unresolved", "summary.md"} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q in finished status; got %q", want, got)
		}
	}
}

// TestComputeStatusLine_FutureEndMTimeFallsThroughToIdle pins the
// regression flagged by critic c1-1 (specs/14, attack-format
// reproduction): a future-dated end.json mtime made `now.Sub(...)`
// negative, which trivially satisfied `<= recentlyDoneWindow` and
// kept the terminal-state line visible indefinitely. The window now
// requires age >= 0 too.
func TestComputeStatusLine_FutureEndMTimeFallsThroughToIdle(t *testing.T) {
	dir := t.TempDir()
	sess := filepath.Join(dir, ".debate", "sessions", "20260507T120000Z-aaa")
	mustMkdir(t, sess)
	endPath := filepath.Join(sess, "end.json")
	mustWrite(t, endPath, `{"termination":{"reason":"steady-state"}}`)

	now := time.Now()
	future := now.Add(24 * time.Hour)
	if err := os.Chtimes(endPath, future, future); err != nil {
		t.Fatal(err)
	}

	if got := computeStatusLine(dir, now); got != idleBar {
		t.Fatalf("future-dated end.json should not be recently done; got %q, want %q", got, idleBar)
	}
}

func TestComputeStatusLine_OldFinishedFallsThroughToIdle(t *testing.T) {
	dir := t.TempDir()
	sess := filepath.Join(dir, ".debate", "sessions", "20260507T120000Z-aaa")
	mustMkdir(t, sess)
	endPath := filepath.Join(sess, "end.json")
	mustWrite(t, endPath, `{}`)
	// Backdate end.json beyond the recently-done window.
	old := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(endPath, old, old); err != nil {
		t.Fatal(err)
	}
	if got := computeStatusLine(dir, time.Now()); got != idleBar {
		t.Errorf("expected idleBar after window, got %q", got)
	}
}

func TestComputeStatusLine_InProgress(t *testing.T) {
	dir := t.TempDir()
	sess := filepath.Join(dir, ".debate", "sessions", "20260507T120000Z-aaa")
	mustMkdir(t, filepath.Join(sess, "forks", "critic-1", "rounds"))
	mustWrite(t, filepath.Join(sess, "start.json"),
		`{"schema":"debate.start.v0","config":{"side_count":4}}`)
	mustWrite(t, filepath.Join(sess, "forks", "critic-1", "rounds", "r1-critic.md"),
		"# Critic 1 - round 1 attacks\n\naspect: input-validation-security\n\n")
	mustWrite(t, filepath.Join(sess, "attacks.jsonl"),
		`{"attack_id":"c1-1","status":"open"}`+"\n"+
			`{"attack_id":"c1-2","status":"conceded"}`+"\n")

	got := computeStatusLine(dir, time.Now().Add(15*time.Second))
	for _, want := range []string{
		"[debate]", "1/4", "input-validation-securi", "R2 proposer", "1o 1c",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q in status line, got %q", want, got)
		}
	}
}

func TestComputeStatusLine_StartingNoRoundsYet(t *testing.T) {
	dir := t.TempDir()
	sess := filepath.Join(dir, ".debate", "sessions", "20260507T120000Z-aaa")
	mustMkdir(t, filepath.Join(sess, "forks", "critic-1", "rounds"))
	mustWrite(t, filepath.Join(sess, "start.json"),
		`{"config":{"side_count":2}}`)
	got := computeStatusLine(dir, time.Now())
	if !strings.Contains(got, "1/2") || !strings.Contains(got, "R1 critic") {
		t.Errorf("expected starting/R1-critic shape, got %q", got)
	}
}

// Regression for critic c1-3 (specs/14): countRoundFiles used to
// match every r*.md file, so a stray markdown file in the rounds
// directory advanced the displayed round count past r1-critic.md.
// Tightened to only accept r<digits>-(critic|proposer).md.
func TestComputeStatusLine_IgnoresNonRoundMarkdown(t *testing.T) {
	dir := t.TempDir()
	sess := filepath.Join(dir, ".debate", "sessions", "20260507T120000Z-aaa")
	rounds := filepath.Join(sess, "forks", "critic-1", "rounds")
	mustMkdir(t, rounds)
	mustWrite(t, filepath.Join(sess, "start.json"), `{"config":{"side_count":1}}`)
	mustWrite(t, filepath.Join(rounds, "r1-critic.md"), "aspect: x\n")
	// A stray non-round markdown file in the same directory must NOT
	// be counted as a round file.
	mustWrite(t, filepath.Join(rounds, "readme.md"), "notes")
	mustWrite(t, filepath.Join(rounds, "r1-critic.md.bak"), "backup")

	got := computeStatusLine(dir, time.Now())
	if !strings.Contains(got, "R2 proposer") {
		t.Errorf("got %q, want status to remain at R2 proposer", got)
	}
}

func TestCountAttackStatuses_FoldByID(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "attacks.jsonl")
	body := `{"attack_id":"c1-1","status":"open"}` + "\n" +
		`{"attack_id":"c1-1","status":"conceded"}` + "\n" +
		`{"attack_id":"c1-2","status":"open"}` + "\n" +
		`{"attack_id":"c1-3","status":"rebutted"}` + "\n"
	mustWrite(t, p, body)
	open, conceded, rebutted := countAttackStatuses(p)
	if open != 1 || conceded != 1 || rebutted != 1 {
		t.Errorf("got open=%d conceded=%d rebutted=%d, want 1/1/1", open, conceded, rebutted)
	}
}

func TestShortTopic(t *testing.T) {
	if got := shortTopic("short"); got != "short" {
		t.Errorf("got %q", got)
	}
	long := strings.Repeat("x", 40)
	got := shortTopic(long)
	if !strings.HasSuffix(got, "…") {
		t.Errorf("long topic not ellipsized: %q", got)
	}
	// 23 ascii xs + the ellipsis rune; rune count is 24, byte length
	// is 23 + 3 (UTF-8 ellipsis is 3 bytes).
	if n := len([]rune(got)); n != 24 {
		t.Errorf("rune count: got %d, want 24", n)
	}
}

func TestFmtElapsed(t *testing.T) {
	cases := map[time.Duration]string{
		5 * time.Second:               "5s",
		59 * time.Second:              "59s",
		60 * time.Second:              "1m00s",
		90 * time.Second:              "1m30s",
		3*time.Minute + 5*time.Second: "3m05s",
	}
	for in, want := range cases {
		if got := fmtElapsed(in); got != want {
			t.Errorf("fmtElapsed(%s) = %q, want %q", in, got, want)
		}
	}
}

// helpers

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
