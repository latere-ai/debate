package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Plan summarizes what one debate run will do; produced by Preflight.
type Plan struct {
	Cwd            string
	Forks          []ForkPlan
	HookMode       bool
	StateDirAbs    string
	SessionID      string
	TranscriptPath string
}

// ForkPlan binds a 1-based critic index to its aspect.
type ForkPlan struct {
	Index  int
	Aspect string
}

// PreflightError carries the exit code the CLI should propagate.
type PreflightError struct {
	Code int
	Msg  string
	Wrap error
}

func (e *PreflightError) Error() string {
	if e.Wrap != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Wrap)
	}
	return e.Msg
}

func (e *PreflightError) Unwrap() error { return e.Wrap }

// ErrRecursionGuard signals "exit 0 immediately"; the orchestrator
// detected DEBATE_IN_PROGRESS in env.
var ErrRecursionGuard = errors.New("recursion guard triggered")

// Preflight runs every pre-flight check against f. On success it
// returns *Plan; on failure a *PreflightError or ErrRecursionGuard.
func Preflight(_ context.Context, f *Flags) (*Plan, error) {
	// 1. Recursion guard — exit 0 fast path.
	if os.Getenv("DEBATE_IN_PROGRESS") != "" {
		return nil, ErrRecursionGuard
	}

	// 2. cwd resolution.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, &PreflightError{Code: 101, Msg: "cannot resolve cwd", Wrap: err}
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return nil, &PreflightError{Code: 101, Msg: "cannot make cwd absolute", Wrap: err}
	}
	if f.Transcript != "" {
		// Best-effort: if the transcript path encodes a different cwd,
		// flag it. The full encoding logic lives in spec 07; here we
		// only short-circuit on the obvious mismatch.
		if expected := decodeCwdFromTranscript(f.Transcript); expected != "" && expected != cwd {
			return nil, &PreflightError{
				Code: 101,
				Msg:  fmt.Sprintf("--transcript points at a session whose cwd does not match the current directory; cd to %s and retry", expected),
			}
		}
	}

	// 3. Env hygiene — never an error; informational only.
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		_ = os.Unsetenv("ANTHROPIC_API_KEY")
		if f.Verbose >= 1 {
			fmt.Fprintln(os.Stderr, "debate: unset stale ANTHROPIC_API_KEY for this run (claude OAuth keychain will be used)")
		}
	}

	// 4. v0 mode gates.
	if f.Main == "codex" {
		return nil, &PreflightError{Code: 102, Msg: "--main codex is v1; v0 supports --main claude only"}
	}
	if f.Judge != "none" {
		return nil, &PreflightError{Code: 103, Msg: "--judge llm/human is v1; v0 supports --judge none only"}
	}

	// 5. Family/model rule.
	if agentFamily(f.Main) == agentFamily(f.Side) {
		if f.MainModel == "" || f.SideModel == "" {
			return nil, &PreflightError{
				Code: 110,
				Msg:  "--main and --side are the same family; both --main-model and --side-model are required and must differ",
			}
		}
		if f.MainModel == f.SideModel {
			return nil, &PreflightError{
				Code: 111,
				Msg:  "--main-model and --side-model must differ when --main and --side are the same family",
			}
		}
	}

	// 6. Side-count vs aspect arity.
	if f.SideCount != len(f.Aspect) {
		return nil, &PreflightError{
			Code: 120,
			Msg:  fmt.Sprintf("--side-count (%d) must equal len(--aspect) (%d)", f.SideCount, len(f.Aspect)),
		}
	}
	if f.SideCount < 1 {
		return nil, &PreflightError{Code: 121, Msg: "--side-count must be ≥ 1"}
	}
	if f.MaxTurn < 2 {
		return nil, &PreflightError{Code: 122, Msg: "--max-turn must be ≥ 2 (one attack + one defense minimum)"}
	}
	if f.CostCap < 1 {
		return nil, &PreflightError{Code: 123, Msg: "--cost-cap must be ≥ 1"}
	}
	if f.ChangedLinesMin < 0 {
		return nil, &PreflightError{Code: 124, Msg: "--changed-lines-min must be ≥ 0"}
	}

	// 7. Task-context source.
	if f.SessionID == "" && f.Transcript == "" && f.TaskContext == "" {
		return nil, &PreflightError{
			Code: 130,
			Msg:  "cannot determine task context: pass --session-id, --transcript, or --task-context",
		}
	}

	// 8. State dir writability.
	stateDir := f.StateDir
	if !filepath.IsAbs(stateDir) {
		stateDir = filepath.Join(cwd, stateDir)
	}
	parent := filepath.Dir(stateDir)
	if info, err := os.Stat(parent); err != nil || !info.IsDir() {
		return nil, &PreflightError{
			Code: 140,
			Msg:  fmt.Sprintf("cannot write under %s", parent),
			Wrap: err,
		}
	}

	// 9. .gitignore advisory — never an error.
	if missingFromGitignore(stateDir) {
		fmt.Fprintln(os.Stderr, "debate: warning: .debate/ is not in .gitignore — consider adding it before committing")
	}

	// Build forks plan.
	forks := make([]ForkPlan, len(f.Aspect))
	for i, a := range f.Aspect {
		forks[i] = ForkPlan{Index: i + 1, Aspect: a}
	}

	return &Plan{
		Cwd:            cwd,
		Forks:          forks,
		HookMode:       f.HookMode,
		StateDirAbs:    stateDir,
		SessionID:      f.SessionID,
		TranscriptPath: f.Transcript,
	}, nil
}

func agentFamily(name string) string {
	switch name {
	case "claude":
		return "claude"
	case "codex":
		return "codex"
	default:
		return name
	}
}

// decodeCwdFromTranscript best-effort decodes the encoded cwd embedded
// in a path like ~/.claude/projects/-Users-x-y/<id>.jsonl. Returns "" if
// the path does not match that shape.
func decodeCwdFromTranscript(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(abs), "/")
	for i, p := range parts {
		if p == "projects" && i+1 < len(parts) {
			encoded := parts[i+1]
			return strings.ReplaceAll(encoded, "-", "/")
		}
	}
	return ""
}

func missingFromGitignore(stateDir string) bool {
	root, err := gitRoot()
	if err != nil {
		return false
	}
	b, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, stateDir)
	if err != nil {
		return false
	}
	target := filepath.ToSlash(rel)
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimSuffix(line, "/")
		if line == target || line == target+"/" || strings.HasPrefix(target+"/", line+"/") {
			return false
		}
	}
	return true
}
