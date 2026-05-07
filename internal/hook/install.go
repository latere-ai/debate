// Package hook owns the install/uninstall logic for the Stop hook
// settings.json entry. See specs/24-stop-hook.md.
package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Scope picks where settings.json lives.
type Scope int

// Scope values.
const (
	// ScopeUser writes ~/.claude/settings.json.
	ScopeUser Scope = iota
	// ScopeProject writes ./.claude/settings.json.
	ScopeProject
)

// SettingsPath returns the absolute settings.json for a scope.
func SettingsPath(scope Scope) (string, error) {
	switch scope {
	case ScopeUser:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude", "settings.json"), nil
	case ScopeProject:
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(cwd, ".claude", "settings.json"), nil
	}
	return "", fmt.Errorf("unknown scope: %d", scope)
}

// Install merges a Stop hook entry into settings.json at scope.
// Idempotent: running twice is the same as running once.
func Install(scope Scope, scriptPath string) error {
	if scriptPath == "" {
		return fmt.Errorf("scriptPath required")
	}
	p, err := SettingsPath(scope)
	if err != nil {
		return err
	}
	settings, err := readSettings(p)
	if err != nil {
		return err
	}
	patched := mergeStopHook(settings, scriptPath)
	return writeSettingsAtomic(p, patched)
}

// Uninstall removes any Stop hook whose command ends in
// debate-stop-hook.sh or `/debate hook`, plus any debate-owned
// statusLine entry.
func Uninstall(scope Scope) error {
	p, err := SettingsPath(scope)
	if err != nil {
		return err
	}
	settings, err := readSettings(p)
	if err != nil {
		return err
	}
	patched := removeStopHook(settings)
	patched = removeStatusLine(patched)
	return writeSettingsAtomic(p, patched)
}

// InstallStatusLine writes a statusLine entry pointing at command.
// Idempotent: replaces an existing debate-owned statusLine; leaves a
// foreign statusLine alone (returns ErrStatusLineConflict so the
// caller can decide to overwrite or skip).
func InstallStatusLine(scope Scope, command string) error {
	if command == "" {
		return fmt.Errorf("statusLine command required")
	}
	p, err := SettingsPath(scope)
	if err != nil {
		return err
	}
	settings, err := readSettings(p)
	if err != nil {
		return err
	}
	if existing, ok := settings["statusLine"].(map[string]any); ok {
		cmd, _ := existing["command"].(string)
		if !statusLineIsDebate(cmd) {
			return ErrStatusLineConflict
		}
	}
	settings["statusLine"] = map[string]any{
		"type":    "command",
		"command": command,
	}
	return writeSettingsAtomic(p, settings)
}

// ErrStatusLineConflict signals that a non-debate statusLine entry is
// already in settings.json; the caller should report the existing
// command rather than overwrite silently.
var ErrStatusLineConflict = fmt.Errorf("statusLine already set to a non-debate command")

func removeStatusLine(settings map[string]any) map[string]any {
	if existing, ok := settings["statusLine"].(map[string]any); ok {
		cmd, _ := existing["command"].(string)
		if statusLineIsDebate(cmd) {
			delete(settings, "statusLine")
		}
	}
	return settings
}

func statusLineIsDebate(cmd string) bool {
	return strings.HasSuffix(cmd, "/debate status") || cmd == "debate status"
}

func readSettings(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

func mergeStopHook(settings map[string]any, scriptPath string) map[string]any {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}
	stopRaw, _ := hooks["Stop"].([]any)

	desired := map[string]any{
		"matcher": "",
		"hooks": []any{
			map[string]any{"type": "command", "command": scriptPath},
		},
	}

	// Filter out any existing entry whose nested command ends in
	// debate-stop-hook.sh, then append our desired one.
	out := []any{}
	for _, item := range stopRaw {
		entry, _ := item.(map[string]any)
		if entry == nil || !entryReferencesDebate(entry) {
			out = append(out, item)
		}
	}
	out = append(out, desired)
	hooks["Stop"] = out
	return settings
}

func removeStopHook(settings map[string]any) map[string]any {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return settings
	}
	stopRaw, _ := hooks["Stop"].([]any)
	out := []any{}
	for _, item := range stopRaw {
		entry, _ := item.(map[string]any)
		if entry == nil || !entryReferencesDebate(entry) {
			out = append(out, item)
		}
	}
	hooks["Stop"] = out
	return settings
}

func entryReferencesDebate(entry map[string]any) bool {
	cmds, _ := entry["hooks"].([]any)
	for _, c := range cmds {
		m, _ := c.(map[string]any)
		if m == nil {
			continue
		}
		s, _ := m["command"].(string)
		// Either the legacy shell-script trampoline or the modern
		// `<path-to>/debate hook` subcommand form. The shell-script
		// form is kept for back-compat with older release tarballs.
		if strings.HasSuffix(s, "debate-stop-hook.sh") {
			return true
		}
		if strings.HasSuffix(s, "/debate hook") || s == "debate hook" {
			return true
		}
	}
	return false
}

func writeSettingsAtomic(path string, settings map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LocateScript returns the path to debate-stop-hook.sh, preferring a
// sibling of the running binary, falling back to one in scripts/ from
// the current cwd.
func LocateScript() string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "debate-stop-hook.sh")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	cwd, err := os.Getwd()
	if err == nil {
		candidate := filepath.Join(cwd, "scripts", "debate-stop-hook.sh")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "debate-stop-hook.sh"
}
