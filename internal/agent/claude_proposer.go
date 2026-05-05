package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ClaudeProposer drives the proposer-clone via claude --resume.
type ClaudeProposer struct {
	Bin      string
	Cwd      string
	RootID   string
	Model    string
	Deadline time.Duration
}

// ProposerResult is one round's outcome.
type ProposerResult struct {
	ForkID       string
	Response     string
	Tokens       int
	USD          float64
	Stdout       []byte
	ChangedFiles []string
	Duration     time.Duration
}

// Typed errors per spec 17.
var (
	ErrCwdMismatch    = errors.New("claude --resume cwd mismatch")
	ErrAuth           = errors.New("claude auth failure")
	ErrTimeout        = errors.New("claude call timed out")
	ErrJSON           = errors.New("claude JSON parse failed")
	ErrEmptyResult    = errors.New("claude returned empty result")
	ErrUnexpectedFork = errors.New("claude session_id mismatch")
	ErrAgentError     = errors.New("claude reported is_error")
)

type claudeJSON struct {
	Type        string  `json:"type"`
	Subtype     string  `json:"subtype"`
	SessionID   string  `json:"session_id"`
	Result      string  `json:"result"`
	IsError     bool    `json:"is_error"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// FirstRound creates a fork and processes R1 in one shot.
func (p *ClaudeProposer) FirstRound(ctx context.Context, pointer string) (*ProposerResult, error) {
	args := []string{"--resume", p.RootID, "--fork-session", "--output-format", "json", "--print", pointer}
	if p.Model != "" {
		args = append(args, "--model", p.Model)
	}
	return p.run(ctx, args, "" /* expected fork id */)
}

// NextRound continues an existing fork.
func (p *ClaudeProposer) NextRound(ctx context.Context, forkID, pointer string) (*ProposerResult, error) {
	args := []string{"--resume", forkID, "--output-format", "json", "--print", pointer}
	if p.Model != "" {
		args = append(args, "--model", p.Model)
	}
	return p.run(ctx, args, forkID)
}

func (p *ClaudeProposer) run(ctx context.Context, args []string, expectFork string) (*ProposerResult, error) {
	bin := p.Bin
	if bin == "" {
		bin = "claude"
	}
	res, err := Exec(ctx, Run{
		Bin: bin, Args: args, Cwd: p.Cwd, Env: CleanEnv(), Deadline: p.Deadline,
	})
	if err != nil {
		if res.Killed {
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		stderr := string(res.Stderr)
		if strings.Contains(stderr, "No conversation found with session ID") {
			return nil, fmt.Errorf("%w: %s", ErrCwdMismatch, stderr)
		}
		if strings.Contains(stderr, "Authentication error") || strings.Contains(stderr, "401") {
			return nil, fmt.Errorf("%w: %s", ErrAuth, stderr)
		}
		return nil, fmt.Errorf("claude exec: %w (stderr=%q)", err, stderr)
	}

	var parsed claudeJSON
	if err := DecodeJSONLine(res.Stdout, &parsed); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSON, err)
	}
	if parsed.IsError {
		return nil, fmt.Errorf("%w: subtype=%q result=%q", ErrAgentError, parsed.Subtype, parsed.Result)
	}
	if parsed.Result == "" {
		return nil, ErrEmptyResult
	}
	if expectFork != "" && parsed.SessionID != expectFork {
		return nil, fmt.Errorf("%w: got %q, want %q", ErrUnexpectedFork, parsed.SessionID, expectFork)
	}
	return &ProposerResult{
		ForkID:   parsed.SessionID,
		Response: parsed.Result,
		Tokens:   parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
		USD:      parsed.TotalCostUSD,
		Stdout:   res.Stdout,
		Duration: res.Duration,
	}, nil
}
