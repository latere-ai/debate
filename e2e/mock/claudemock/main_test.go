package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func fakeRead(content string, err error) func(string) ([]byte, error) {
	return func(string) ([]byte, error) {
		if err != nil {
			return nil, err
		}
		return []byte(content), nil
	}
}

func TestDecideDefaultWhenNoScript(t *testing.T) {
	stdout, stderr, code, err := decide("--print x", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("code: got %d, want 0", code)
	}
	if stderr != "" {
		t.Errorf("stderr: got %q", stderr)
	}
	if stdout["type"] != "result" || stdout["subtype"] != "success" {
		t.Errorf("default stdout shape wrong: %v", stdout)
	}
}

func TestDecideMatchingRule(t *testing.T) {
	script := `[
		{"args_contains": "--fork-session", "stdout": {"result": "fork-result"}, "exit": 0},
		{"stdout": {"result": "fallback"}, "exit": 0}
	]`
	stdout, _, code, err := decide("--resume X --fork-session -p y", "/path", fakeRead(script, nil))
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 || stdout["result"] != "fork-result" {
		t.Errorf("matching rule failed: code=%d stdout=%v", code, stdout)
	}

	stdout, _, code, err = decide("--print y", "/path", fakeRead(script, nil))
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 || stdout["result"] != "fallback" {
		t.Errorf("fallback rule failed: code=%d stdout=%v", code, stdout)
	}
}

func TestDecideStderrAndCode(t *testing.T) {
	script := `[{"stdout": null, "stderr": "boom\n", "exit": 7}]`
	stdout, stderr, code, err := decide("anything", "/path", fakeRead(script, nil))
	if err != nil {
		t.Fatal(err)
	}
	if code != 7 {
		t.Errorf("code: got %d, want 7", code)
	}
	if stderr != "boom\n" {
		t.Errorf("stderr: got %q", stderr)
	}
	if stdout != nil {
		t.Errorf("stdout: got %v, want nil", stdout)
	}
}

func TestDecideNoRuleMatched(t *testing.T) {
	script := `[{"args_contains": "never-matches", "stdout": null, "exit": 0}]`
	stdout, stderr, code, err := decide("real args", "/path", fakeRead(script, nil))
	if err != nil {
		t.Fatal(err)
	}
	if code != 3 {
		t.Errorf("code: got %d, want 3", code)
	}
	if stdout != nil {
		t.Errorf("stdout should be nil")
	}
	if !strings.Contains(stderr, "no rule matched") {
		t.Errorf("stderr should mention no rule matched: %q", stderr)
	}
}

func TestDecideReadError(t *testing.T) {
	if _, _, _, err := decide("x", "/nope", fakeRead("", errors.New("file gone"))); err == nil {
		t.Error("expected read error")
	}
}

func TestDecideBadJSON(t *testing.T) {
	if _, _, _, err := decide("x", "/p", fakeRead("{not-json", nil)); err == nil {
		t.Error("expected parse error")
	}
}

func TestEmitWritesStdout(t *testing.T) {
	var out, errw bytes.Buffer
	emit(&out, &errw, map[string]any{"k": "v"}, "", 0)
	if !strings.Contains(out.String(), `"v"`) {
		t.Errorf("stdout missing value: %q", out.String())
	}
}

func TestEmitWritesStderr(t *testing.T) {
	var out, errw bytes.Buffer
	emit(&out, &errw, nil, "warn\n", 0)
	if errw.String() != "warn\n" {
		t.Errorf("stderr: got %q", errw.String())
	}
	if out.String() != "" {
		t.Errorf("nil stdout should write nothing: %q", out.String())
	}
}
