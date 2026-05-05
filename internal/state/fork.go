package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Role discriminates critic (odd rounds) from proposer (even rounds).
type Role int

const (
	RoleCritic Role = iota
	RoleProposer
)

// ProposerState mirrors specs/11-fork-artifacts.md. Two shapes,
// discriminated by Agent.
type ProposerState struct {
	Schema          string          `json:"schema"`
	Agent           string          `json:"agent"`
	Model           string          `json:"model,omitempty"`
	ForkSessionID   string          `json:"fork_session_id,omitempty"`
	RootSessionID   string          `json:"root_session_id,omitempty"`
	RoundThreadIDs  []RoundThreadID `json:"round_thread_ids,omitempty"`
}

// RoundThreadID records one codex thread per even round.
type RoundThreadID struct {
	Round    int    `json:"round"`
	ThreadID string `json:"thread_id"`
}

const SchemaProposerState = "debate.proposer-state.v0"

// ErrRoleParity is returned when round parity disagrees with role.
var ErrRoleParity = errors.New("round/role parity mismatch")

// WriteProposerState writes <session>/forks/critic-<i>/proposer-state.json.
func WriteProposerState(s *Session, fork int, ps *ProposerState) error {
	if ps.Schema == "" {
		ps.Schema = SchemaProposerState
	}
	b, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return err
	}
	return s.AtomicWrite(forkPath(fork, "proposer-state.json"), append(b, '\n'))
}

// WriteForkDiff writes per-fork diff.patch (what this critic actually saw).
func WriteForkDiff(s *Session, fork int, patch string) error {
	return s.AtomicWrite(forkPath(fork, "diff.patch"), []byte(patch))
}

// WriteRunDiff writes <session>/diff.patch (the run-level initial snapshot).
func WriteRunDiff(s *Session, patch string) error {
	return s.AtomicWrite("diff.patch", []byte(patch))
}

// WriteRound writes <session>/forks/critic-<i>/rounds/r<n>-<role>.md.
// Round parity is enforced: critic on odd, proposer on even.
func WriteRound(s *Session, fork, round int, role Role, body []byte) error {
	if (role == RoleCritic && round%2 == 0) || (role == RoleProposer && round%2 == 1) {
		return ErrRoleParity
	}
	name := fmt.Sprintf("r%d-%s.md", round, roleName(role))
	return s.AtomicWrite(forkPath(fork, "rounds", name), body)
}

// ChangedFilesAfter returns paths modified in cwd since `since` was
// captured (or, when since is nil, every tracked-modified file vs HEAD
// plus untracked).
func ChangedFilesAfter(ctx context.Context, cwd string, since []string) ([]string, error) {
	out, err := exec.CommandContext(ctx, "git", "status", "--porcelain").CombinedOutput()
	if err != nil {
		return nil, err
	}
	now := parsePorcelain(string(out))
	prior := map[string]bool{}
	for _, p := range since {
		prior[p] = true
	}
	added := make([]string, 0, len(now))
	for _, p := range now {
		if !prior[p] {
			added = append(added, p)
		}
	}
	return added, nil
}

func parsePorcelain(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if len(line) < 4 {
			continue
		}
		out = append(out, strings.TrimSpace(line[3:]))
	}
	return out
}

func forkPath(fork int, parts ...string) string {
	all := append([]string{"forks", fmt.Sprintf("critic-%d", fork)}, parts...)
	return filepath.Join(all...)
}

func roleName(r Role) string {
	if r == RoleCritic {
		return "critic"
	}
	return "proposer"
}
