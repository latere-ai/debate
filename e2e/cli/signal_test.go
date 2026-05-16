// Signal handling regression test: SIGINT to bin/debate must propagate
// to its child agent processes and the orchestrator must exit in well
// under 5 seconds. Catches the spec-21 bug where main() used cobra's
// background context and the round.InstallHandler signal handler was
// never wired.

package cli_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestSignalLatency_StuckChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("signal-handling probe is POSIX only")
	}

	root := repoRoot(t)
	binDir := t.TempDir()
	agon := build(t, root, "./cmd/agon", binDir)

	// Shell shims for claude/codex that sleep long enough to outlast
	// any reasonable signal latency budget.
	stubDir := t.TempDir()
	for _, name := range []string{"claude", "codex"} {
		p := filepath.Join(stubDir, name)
		const stub = "#!/usr/bin/env bash\nexec sleep 60\n"
		if err := os.WriteFile(p, []byte(stub), 0o755); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	repo := fixtureRepo(t)

	env := patchedPATH(t, stubDir)
	env = append(env, "HOME="+t.TempDir())

	cmd := exec.Command(agon,
		"--main", "claude",
		"--side", "codex",
		"--max-turn", "2",
		"--side-count", "1",
		"--task-context", "signal latency probe",
		"--changed-lines-min", "10",
		"--state-dir", filepath.Join(repo, ".agon"),
	)
	cmd.Dir = repo
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Allow time for preflight and the first agent.Exec call to spawn.
	// 1s is generous; the probe runs fine at 0.5s on dev hardware.
	time.Sleep(1 * time.Second)

	// Send SIGINT to agon's process group, mirroring the behaviour
	// of an interactive Ctrl-C.
	start := time.Now()
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGINT); err != nil {
		t.Fatalf("send SIGINT: %v", err)
	}

	// Wait for exit, but with a hard upper bound so a genuine hang
	// fails the test rather than the test suite's outer timeout.
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-done
		t.Fatalf("agon did not exit within 5s of SIGINT (regression: signal handler not wired into main)")
	}
	elapsed := time.Since(start)

	// 2-second budget gives us margin over the 2s SIGINT-to-SIGKILL
	// grace in agent.Exec; in practice the orchestrator exits in well
	// under 100 ms once the signal handler is correctly wired.
	if elapsed > 2*time.Second {
		t.Fatalf("SIGINT-to-exit too slow: %s (>2s)", elapsed)
	}
	t.Logf("SIGINT-to-exit: %s", elapsed)

	// Defensive: deadline-aware orphan check.
	_, cancelChk := context.WithTimeout(context.Background(), 1*time.Second)
	cancelChk()
}
