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

// TokenUsage captures the per-call token breakdown reported by claude's
// `--output-format json`. Tokens reports a single billed-input total
// (Input + CacheCreate + CacheRead) plus Output. The fields stay
// separate so callers can show prompt-cache hit rate or estimate cost.
type TokenUsage struct {
	Input       int `json:"input_tokens"`
	Output      int `json:"output_tokens"`
	CacheCreate int `json:"cache_creation_input_tokens"`
	CacheRead   int `json:"cache_read_input_tokens"`
}

// Total returns the sum of every token bucket. Useful as a single
// cost-cap dial; for billing accuracy use the individual fields and the
// model's per-bucket price.
func (u TokenUsage) Total() int {
	return u.Input + u.Output + u.CacheCreate + u.CacheRead
}

// Add accumulates other into u in place.
func (u *TokenUsage) Add(other TokenUsage) {
	u.Input += other.Input
	u.Output += other.Output
	u.CacheCreate += other.CacheCreate
	u.CacheRead += other.CacheRead
}

// ProposerResult is one round's outcome.
type ProposerResult struct {
	ForkID       string
	Response     string
	Tokens       int
	Usage        TokenUsage
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
	Type         string  `json:"type"`
	Subtype      string  `json:"subtype"`
	SessionID    string  `json:"session_id"`
	Result       string  `json:"result"`
	IsError      bool    `json:"is_error"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	Usage        struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

// usage extracts the typed TokenUsage from a parsed claude JSON
// document. Helper to keep callsites readable.
func (j *claudeJSON) usage() TokenUsage {
	return TokenUsage{
		Input:       j.Usage.InputTokens,
		Output:      j.Usage.OutputTokens,
		CacheCreate: j.Usage.CacheCreationInputTokens,
		CacheRead:   j.Usage.CacheReadInputTokens,
	}
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
	use := parsed.usage()
	return &ProposerResult{
		ForkID:   parsed.SessionID,
		Response: parsed.Result,
		Tokens:   use.Input + use.Output,
		Usage:    use,
		USD:      parsed.TotalCostUSD,
		Stdout:   res.Stdout,
		Duration: res.Duration,
	}, nil
}
