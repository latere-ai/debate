package input

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEncodeDecodeRoundtrip(t *testing.T) {
	// Note: claude's encoding (literal '/' -> '-') is lossy when the
	// cwd contains '-'; we mirror that and only round-trip cases
	// without ambiguity.
	cases := []string{
		"/Users/changkun/dev/foo",
		"/srv/something/x",
		"/",
	}
	for _, c := range cases {
		got := DecodeCwd(EncodeCwd(c))
		if got != c {
			t.Errorf("Encode/Decode round-trip: got %q, want %q", got, c)
		}
	}
}

func TestExtractFirstUserStringForm(t *testing.T) {
	rec := []byte(`{"type":"user","message":{"role":"user","content":"do the task"}}`)
	got, err := ExtractFirstUser([][]byte{rec})
	if err != nil {
		t.Fatal(err)
	}
	if got != "do the task" {
		t.Errorf("got %q", got)
	}
}

func TestExtractFirstUserArrayForm(t *testing.T) {
	rec := []byte(`{"type":"user","message":{"role":"user","content":[{"type":"text","text":"a"},{"type":"text","text":"b"}]}}`)
	got, err := ExtractFirstUser([][]byte{rec})
	if err != nil {
		t.Fatal(err)
	}
	if got != "a\n\nb" {
		t.Errorf("got %q", got)
	}
}

func TestReadTranscriptMissing(t *testing.T) {
	_, err := ReadTranscript(filepath.Join(t.TempDir(), "missing.jsonl"))
	if err == nil {
		t.Fatal("want error for missing file")
	}
}

func TestReadTranscriptNoUserTurn(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "abc.jsonl")
	if err := os.WriteFile(p, []byte(`{"type":"system","message":{"role":"system","content":"x"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadTranscript(p)
	if !errors.Is(err, ErrNoUserTurn) {
		t.Errorf("got %v, want ErrNoUserTurn", err)
	}
}

func TestReadTranscriptHappy(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "abc.jsonl")
	body := `{"type":"system","message":{"role":"system","content":"sys"}}` + "\n" +
		`{"type":"user","message":{"role":"user","content":"do the task"},"timestamp":"2026-05-06T14:12:33Z"}` + "\n" +
		`{"type":"assistant","message":{"role":"assistant","content":"ok"}}` + "\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	tr, err := ReadTranscript(p)
	if err != nil {
		t.Fatal(err)
	}
	if tr.FirstUser != "do the task" {
		t.Errorf("FirstUser: got %q", tr.FirstUser)
	}
	if tr.SessionID != "abc" {
		t.Errorf("SessionID: got %q", tr.SessionID)
	}
}

func TestLocateTranscriptExplicit(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.jsonl")
	if err := os.WriteFile(p, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LocateTranscript("", "", "", p)
	if err != nil {
		t.Fatal(err)
	}
	if got != p {
		t.Errorf("got %q", got)
	}
}
