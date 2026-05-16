package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// roundFileRE matches the on-disk round-file naming contract:
// r<n>-critic.md or r<n>-proposer.md (n >= 1). Matching anything
// looser (e.g. `r*.md`) lets a stray markdown file in the rounds
// directory advance the displayed round count.
var roundFileRE = regexp.MustCompile(`^r\d+-(critic|proposer)\.md$`)

// statusCmd renders a one-line status for the most recent agon run
// in the current cwd. Designed for Claude Code's `statusLine` setting:
// the binary is invoked every few seconds, must exit fast, and prints
// one line of stdout that the TUI renders at the bottom of the
// session.
//
// Empty output means "nothing to show" (no agon state in cwd, or
// the latest run is already finished). statusLine treats empty stdout
// as "no status bar," so absence is a clean no-op.
func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "status",
		Short:  "One-line status for the most recent agon run (statusLine integration)",
		Hidden: false,
		RunE: func(_ *cobra.Command, _ []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return nil // statusLine must never error; print nothing
			}
			line := computeStatusLine(cwd, time.Now())
			if line != "" {
				fmt.Println(line)
			}
			return nil
		},
	}
}

// recentlyDoneWindow controls how long after a session finishes we
// keep showing the terminal-state line in the status bar instead of
// going silent. Long enough that a finished run is visible after the
// user notices their prompt returned; short enough not to confuse a
// later, brand-new run.
const recentlyDoneWindow = 2 * time.Minute

// idleBar is what we print when there's nothing to show. A single
// space rather than the empty string: claude's statusLine keeps the
// previous frame visible across empty outputs, which would freeze
// the bar on the last in-flight snapshot of a finished run. A space
// is non-empty so claude renders it and the stale frame goes away.
const idleBar = " "

// computeStatusLine inspects <cwd>/.agon/sessions/<latest>/ and
// returns a concise progress line. now is injected for tests; pass
// time.Now() in production.
func computeStatusLine(cwd string, now time.Time) string {
	sessionsDir := filepath.Join(cwd, ".agon", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil || len(entries) == 0 {
		return idleBar
	}

	// Sessions are named "<UTC-timestamp>-<rand>" so lexicographic max
	// is also temporal max. Filter to dirs only just in case.
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return idleBar
	}
	sort.Strings(names)
	sessDir := filepath.Join(sessionsDir, names[len(names)-1])

	// Finished. Show a terminal-state summary briefly so the user can
	// see "the run just ended"; after that, fall through to idleBar so
	// the stale progress frame doesn't stick.
	//
	// Require age >= 0 too: a future-dated end.json (clock skew,
	// restored backup, or fixture-generated future mtime) would
	// otherwise produce a negative `now.Sub(...)` that trivially
	// satisfies `<= window` and pin the terminal-state line forever.
	if endInfo, err := os.Stat(filepath.Join(sessDir, "end.json")); err == nil {
		age := now.Sub(endInfo.ModTime())
		if age >= 0 && age <= recentlyDoneWindow {
			return finishedStatusLine(sessDir)
		}
		return idleBar
	}

	// Total fork count from start.json (config.side_count).
	totalForks := 0
	if b, err := os.ReadFile(filepath.Join(sessDir, "start.json")); err == nil {
		var start struct {
			Config struct {
				SideCount int `json:"side_count"`
			} `json:"config"`
		}
		_ = json.Unmarshal(b, &start)
		totalForks = start.Config.SideCount
	}

	// Find which fork has the most-recently-written round file. In v0
	// forks run serially, so newest mtime == active fork.
	forksDir := filepath.Join(sessDir, "forks")
	activeIdx, activeName, newestMtime := findActiveFork(forksDir)
	if activeName == "" {
		return fmt.Sprintf("[agon] starting (1/%d)", max(totalForks, 1))
	}

	roundsDir := filepath.Join(forksDir, activeName, "rounds")
	roundFiles := countRoundFiles(roundsDir)

	// Round naming: r1-critic.md, r2-proposer.md, r3-critic.md, ...
	// The next round to write is roundFiles+1.
	nextRound := roundFiles + 1
	role := "critic"
	if nextRound%2 == 0 {
		role = "proposer"
	}

	// Topic from r1-critic.md aspect: line, if it exists yet.
	topic := "topic pending"
	if t := readTopic(filepath.Join(roundsDir, "r1-critic.md")); t != "" {
		topic = t
	}

	// Elapsed since the most recent round file landed (i.e. how long
	// the current phase has been running). Falls back to session
	// start.json mtime if we couldn't find a round file yet.
	if newestMtime.IsZero() {
		if info, err := os.Stat(filepath.Join(sessDir, "start.json")); err == nil {
			newestMtime = info.ModTime()
		}
	}
	elapsed := now.Sub(newestMtime).Round(time.Second)
	if elapsed < 0 {
		elapsed = 0
	}

	open, conceded, rebutted := countAttackStatuses(filepath.Join(sessDir, "attacks.jsonl"))

	line := fmt.Sprintf("[agon] %d/%d %s | R%d %s %s",
		activeIdx, max(totalForks, activeIdx),
		shortTopic(topic), nextRound, role, fmtElapsed(elapsed))
	if open+conceded+rebutted > 0 {
		line += fmt.Sprintf(" | %do %dc %dr", open, conceded, rebutted)
	}
	return line
}

// finishedStatusLine renders a one-line "the run just ended" summary
// from end.json: termination reason + unresolved/total counts. Falls
// back to a generic "[agon] done" if end.json is unparseable.
func finishedStatusLine(sessDir string) string {
	b, err := os.ReadFile(filepath.Join(sessDir, "end.json"))
	if err != nil {
		return "[agon] done"
	}
	var end struct {
		Termination struct {
			Reason string `json:"reason"`
		} `json:"termination"`
		Stats struct {
			ByStatus map[string]int `json:"by_status"`
			Total    int            `json:"total_attacks"`
		} `json:"stats"`
	}
	if err := json.Unmarshal(b, &end); err != nil {
		return "[agon] done"
	}
	reason := end.Termination.Reason
	if reason == "" {
		reason = "done"
	}
	unresolved := end.Stats.ByStatus["unresolved"] + end.Stats.ByStatus["open"]
	if end.Stats.Total > 0 {
		return fmt.Sprintf("[agon] %s • %d/%d unresolved • see summary.md",
			reason, unresolved, end.Stats.Total)
	}
	return fmt.Sprintf("[agon] %s", reason)
}

func findActiveFork(forksDir string) (idx int, name string, newest time.Time) {
	entries, err := os.ReadDir(forksDir)
	if err != nil {
		return 0, "", time.Time{}
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		var i int
		if _, err := fmt.Sscanf(e.Name(), "critic-%d", &i); err != nil {
			continue
		}
		rd := filepath.Join(forksDir, e.Name(), "rounds")
		rf, err := os.ReadDir(rd)
		if err != nil {
			continue
		}
		for _, r := range rf {
			info, err := r.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(newest) {
				newest = info.ModTime()
				name = e.Name()
				idx = i
			}
		}
	}
	if name == "" {
		// No round files anywhere yet; pick critic-1 as the active
		// candidate if its dir exists, so the line still renders
		// "1/N starting" instead of empty.
		for _, e := range entries {
			if e.IsDir() && e.Name() == "critic-1" {
				return 1, "critic-1", time.Time{}
			}
		}
	}
	return idx, name, newest
}

func countRoundFiles(roundsDir string) int {
	entries, err := os.ReadDir(roundsDir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if roundFileRE.MatchString(e.Name()) {
			n++
		}
	}
	return n
}

func readTopic(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	// 4 KB is enough; the aspect: line is in the first dozen lines.
	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	for line := range strings.SplitSeq(string(buf[:n]), "\n") {
		if strings.HasPrefix(line, "aspect:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "aspect:"))
		}
	}
	return ""
}

func countAttackStatuses(path string) (open, conceded, rebutted int) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, 0
	}
	// attacks.jsonl is append-only; later entries supersede earlier
	// for the same attack_id. Fold by id, then count.
	type rec struct {
		AttackID string `json:"attack_id"`
		Status   string `json:"status"`
	}
	state := map[string]string{}
	for line := range strings.SplitSeq(string(b), "\n") {
		if line == "" {
			continue
		}
		var r rec
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			continue
		}
		if r.AttackID != "" {
			state[r.AttackID] = r.Status
		}
	}
	for _, s := range state {
		switch s {
		case "open", "unresolved":
			open++
		case "conceded":
			conceded++
		case "rebutted", "withdrawn":
			rebutted++
		}
	}
	return open, conceded, rebutted
}

func shortTopic(t string) string {
	const limit = 24
	if len(t) <= limit {
		return t
	}
	return t[:limit-1] + "…"
}

func fmtElapsed(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) - m*60
	return fmt.Sprintf("%dm%02ds", m, s)
}

// _ ensures fs is referenced for the build (kept for future use of
// ReadDir typed errors).
var _ fs.FileMode = 0
