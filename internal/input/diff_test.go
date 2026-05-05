package input

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"-c", "user.email=t@e.com", "-c", "user.name=t", "commit", "--allow-empty", "-q", "-m", "init"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

func TestComputeEmptyDiff(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	d, err := Compute(context.Background(), DiffSpec{From: "HEAD", To: ".", Cwd: dir})
	if err != nil {
		t.Fatal(err)
	}
	if d.ChangedLines != 0 {
		t.Errorf("ChangedLines: got %d, want 0", d.ChangedLines)
	}
}

func TestComputeWithUntracked(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := Compute(context.Background(), DiffSpec{From: "HEAD", To: ".", Cwd: dir})
	if err != nil {
		t.Fatal(err)
	}
	if d.ChangedLines == 0 {
		t.Errorf("expected non-zero ChangedLines for untracked file; patch:\n%s", d.Patch)
	}
}

func TestTrivialGate(t *testing.T) {
	d := &Diff{ChangedLines: 3}
	if !Trivial(d, 10) {
		t.Error("expected trivial")
	}
	if Trivial(d, 2) {
		t.Error("expected non-trivial")
	}
}

func TestComputeNotARepo(t *testing.T) {
	dir := t.TempDir()
	_, err := Compute(context.Background(), DiffSpec{From: "HEAD", To: ".", Cwd: dir})
	if err == nil {
		t.Fatal("expected ErrNotGitRepo")
	}
}
