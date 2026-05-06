package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"latere.ai/x/debate/internal/critic"
)

// Critic is the interface every critic driver satisfies.
type Critic interface {
	Round(ctx context.Context, in CriticInput) (*CriticResult, error)
}

// CriticInput parameterizes one critic round.
type CriticInput struct {
	Aspect          critic.Aspect
	CriticIndex     int
	Round           int
	SystemPrompt    string
	TaskContext     string
	DiffPatch       string
	PriorRoundFiles []RoundFileRef
	Cwd             string
	Deadline        time.Duration
	Model           string
}

// RoundFileRef points at a prior round file for the critic to read.
type RoundFileRef struct {
	Path  string
	Round int
	Role  string
}

// CriticResult is one round's outcome.
type CriticResult struct {
	Markdown string
	ThreadID string
	Tokens   int
	USD      float64
	Stdout   []byte
	Duration time.Duration
}

// Typed errors per spec 18.
var (
	ErrRateLimit    = errors.New("agent reported rate limit")
	ErrTTYRequired  = errors.New("agent requires a TTY")
	ErrEmptyContent = errors.New("agent returned empty content")
)

// NewCritic returns a Critic for the named family. Panics on unknown.
func NewCritic(family string) Critic {
	switch family {
	case "codex":
		return &CodexCritic{}
	case "claude":
		return &ClaudeCritic{}
	default:
		panic("unknown critic family: " + family)
	}
}

// AssemblePrompt is the deterministic prompt a critic driver feeds to
// the agent: aspect system prompt + task + diff + pointers to prior
// rounds.
func AssemblePrompt(in CriticInput) string {
	var b strings.Builder
	b.WriteString(in.SystemPrompt)
	b.WriteString("\n\n# Task\n\n")
	b.WriteString(in.TaskContext)
	b.WriteString("\n\n# Diff\n\n```diff\n")
	b.WriteString(in.DiffPatch)
	b.WriteString("\n```\n")
	if len(in.PriorRoundFiles) > 0 {
		b.WriteString("\n# Prior rounds\n\n")
		for _, r := range in.PriorRoundFiles {
			fmt.Fprintf(&b, "- @%s — round %d %s\n", r.Path, r.Round, r.Role)
		}
	}
	return b.String()
}

// CodexCritic invokes `codex exec --sandbox read-only --json`.
type CodexCritic struct {
	Bin string
}

// Round runs one codex critic round.
func (c *CodexCritic) Round(ctx context.Context, in CriticInput) (*CriticResult, error) {
	bin := c.Bin
	if bin == "" {
		bin = "codex"
	}
	prompt := AssemblePrompt(in)
	args := []string{"exec", "--skip-git-repo-check", "--sandbox", "read-only", "--json", prompt}
	if in.Model != "" {
		args = append(args, "--model", in.Model)
	}
	res, err := Exec(ctx, Run{
		Bin: bin, Args: args, Cwd: in.Cwd, Env: CleanEnv(), Deadline: in.Deadline,
	})
	if err != nil {
		stderr := string(res.Stderr)
		switch {
		case res.Killed:
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		case strings.Contains(stderr, "rate limit"):
			return nil, fmt.Errorf("%w: %s", ErrRateLimit, stderr)
		case strings.Contains(stderr, "stdin is not a terminal"):
			return nil, fmt.Errorf("%w: %s", ErrTTYRequired, stderr)
		}
		return nil, fmt.Errorf("codex exec: %w (stderr=%q)", err, stderr)
	}

	var (
		out      strings.Builder
		threadID string
	)
	visit := func(raw json.RawMessage) error {
		var ev struct {
			Type     string `json:"type"`
			ThreadID string `json:"thread_id"`
			Item     struct {
				Type    string `json:"type"`
				Content string `json:"content"`
			} `json:"item"`
		}
		if err := json.Unmarshal(raw, &ev); err != nil {
			return nil
		}
		switch ev.Type {
		case "thread.started":
			threadID = ev.ThreadID
		case "item.completed":
			if ev.Item.Type == "agent_message" {
				if out.Len() > 0 {
					out.WriteString("\n")
				}
				out.WriteString(ev.Item.Content)
			}
		}
		return nil
	}
	if err := StreamJSON(strings.NewReader(string(res.Stdout)), visit); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSON, err)
	}
	if out.Len() == 0 {
		return nil, ErrEmptyContent
	}
	return &CriticResult{
		Markdown: out.String(),
		ThreadID: threadID,
		Stdout:   res.Stdout,
		Duration: res.Duration,
	}, nil
}

// ClaudeCritic invokes a fresh `claude -p` per round (no --resume,
// no --fork-session — see spec 18).
type ClaudeCritic struct {
	Bin string
}

// Round runs one claude critic round.
func (c *ClaudeCritic) Round(ctx context.Context, in CriticInput) (*CriticResult, error) {
	bin := c.Bin
	if bin == "" {
		bin = "claude"
	}
	prompt := AssemblePrompt(in)
	args := []string{"--output-format", "json", "--print", prompt}
	if in.Model != "" {
		args = append(args, "--model", in.Model)
	}
	res, err := Exec(ctx, Run{
		Bin: bin, Args: args, Cwd: in.Cwd, Env: CleanEnv(), Deadline: in.Deadline,
	})
	if err != nil {
		if res.Killed {
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		return nil, fmt.Errorf("claude exec: %w (stderr=%q)", err, string(res.Stderr))
	}
	var parsed claudeJSON
	if err := DecodeJSONLine(res.Stdout, &parsed); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSON, err)
	}
	if parsed.IsError {
		return nil, fmt.Errorf("%w: subtype=%q", ErrAgentError, parsed.Subtype)
	}
	if parsed.Result == "" {
		return nil, ErrEmptyContent
	}
	return &CriticResult{
		Markdown: parsed.Result,
		ThreadID: parsed.SessionID,
		Tokens:   parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
		USD:      parsed.TotalCostUSD,
		Stdout:   res.Stdout,
		Duration: res.Duration,
	}, nil
}
