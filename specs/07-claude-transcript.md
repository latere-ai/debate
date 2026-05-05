# Spec 07 — Claude session JSONL reader

> **Status: ✅ implemented.**
> Implementation spec for `debate`. See [01-overview.md](01-overview.md) §"Verified primitives" for design intent.

**Depends on:** [02](02-go-module.md), [04](04-cli-flags.md), [06](06-preflight.md).
**Consumed by:** [10](10-run-artifacts.md), [17](17-claude-proposer.md), [19](19-round-loop.md).

## Scope

In: locating the root session's JSONL, parsing it as the line-delimited JSON Claude Code writes, extracting task context (the first user turn) and basic metadata. Includes a fallback for the case where the path can't be located but `--session-id` is set.

Out: writing or modifying any session file (root is read-only); transcript indexing for the debate's own state-dir (that's [10](10-run-artifacts.md)'s `transcript.jsonl`).

## Path encoding

Claude Code stores root sessions at:

```
$HOME/.claude/projects/<encoded-cwd>/<session-id>.jsonl
```

Encoding rule (verified against claude 2.1.129): replace `/` with `-` in the absolute cwd, leading `-` preserved. Example: `/Users/changkun/dev/foo` becomes `-Users-changkun-dev-foo`. No percent-encoding, no other character substitution.

`Encode(cwd) string` and `Decode(encoded) string` both round-trip.

## Public Go interfaces

```go
// internal/input/transcript.go
package input

type Transcript struct {
    Path        string         // absolute path to <id>.jsonl
    SessionID   string         // matches the basename
    Cwd         string         // decoded from the projects/ segment
    FirstUser   string         // first user turn's content (extracted task context)
    StartedAt   time.Time      // timestamp on the first record
    LineCount   int            // total records read
}

// LocateTranscript resolves the on-disk path for a root session.
//
// Preference order:
//   1. Flags.Transcript (explicit path)
//   2. ~/.claude/projects/<encoded-cwd>/<flags.SessionID>.jsonl
//
// Returns ErrTranscriptNotFound when neither yields an existing file.
func LocateTranscript(home, cwd, sessionID, explicit string) (string, error)

// ReadTranscript opens a JSONL transcript, parses every record, and
// returns a *Transcript with the fields above populated.
//
// Reads are streaming (line-by-line) to keep memory bounded for long
// sessions. Records are decoded into json.RawMessage; only the fields
// this spec needs are extracted.
func ReadTranscript(path string) (*Transcript, error)

// ExtractFirstUser walks records in order and returns the .message.content
// of the first record whose .type == "user". Multi-part user messages
// (string array of role-tagged parts) are joined with "\n\n".
func ExtractFirstUser(records [][]byte) (string, error)
```

Errors:

- `ErrTranscriptNotFound` — wraps `os.ErrNotExist`; carries the path searched.
- `ErrTranscriptMalformed` — wraps the line number and underlying JSON error.
- `ErrNoUserTurn` — transcript contains zero records of `type == "user"`.

## Record shape

Each line is a JSON object. Relevant fields (the rest are passed through as `json.RawMessage` and ignored):

```jsonc
{
  "type": "user" | "assistant" | "system" | "tool_use" | "tool_result" | "hook_*",
  "message": {
    "role": "user" | "assistant" | "system",
    "content": "<string>" | [
      {"type": "text", "text": "<string>"},
      {"type": "tool_use", ...},
      ...
    ]
  },
  "timestamp": "<RFC3339>",
  "uuid": "<message id>",
  ...
}
```

`ExtractFirstUser` handles both the string and array shapes for `content`. Tool-use parts are skipped; only `text` parts are joined.

## Fallback path

If `LocateTranscript` fails but `flags.SessionID != ""`, [17](17-claude-proposer.md) can still drive the proposer (claude knows the session by id). For task-context extraction, the fallback is:

```
claude --resume <session-id> --print "Reply with the original task description, verbatim, no commentary."
```

This is a last resort; the orchestrator emits a stderr warning at `--verbose >= 1` because each invocation spends ~$0.20 priming the system-prompt cache (see [01-overview.md](01-overview.md) §"First-call cost"). Implemented behind:

```go
func TaskContextViaResume(ctx context.Context, sessionID string) (string, error)
```

`TaskContextViaResume` is *not* called from pre-flight; only [10](10-run-artifacts.md)/[19](19-round-loop.md) invoke it after pre-flight succeeds.

## Behavior

- `cwd` decoded from the path is canonicalized (trailing slash stripped, `..` resolved).
- Lines that don't parse as JSON are skipped at `--verbose >= 2` and counted; `Transcript.LineCount` reflects valid records only. A transcript with > 5% bad lines triggers `ErrTranscriptMalformed` (defensive: the file is probably corrupted or a different format).
- Control characters in record fields are tolerated (matches the [01-overview.md](01-overview.md) §"JSON output may contain control characters" finding); the parser uses `encoding/json` directly, never `jq`-style shell pipes.
- File is opened read-only; never modified.

## Test contract

- Round-trip: `Encode(Decode(x)) == x` for a table of cwd strings (incl. `/`, `/Users/me`, `/srv/space dir/`).
- Fixture transcript with a known first user turn → `ExtractFirstUser` returns it verbatim.
- Multi-part user turn (array form) → joined with `\n\n`.
- Transcript missing on disk → `ErrTranscriptNotFound`.
- Transcript with no user turn → `ErrNoUserTurn`.
- Transcript with control chars in `content` parses cleanly.

Fixtures live under `testdata/transcripts/`.

## Acceptance criteria

- [x] `LocateTranscript` handles the four input combinations of `(transcript, sessionID)` defined in the preference order.
- [x] `ReadTranscript` is streaming (verified by reading a 50MB synthetic transcript with `MaxRSS < 50MB`).
- [x] All three error types are reachable in tests with their carried metadata.
- [x] Fallback `TaskContextViaResume` exists but is not used unless explicitly invoked; lint check ensures pre-flight does not import it.
