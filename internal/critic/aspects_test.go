package critic

import (
	"strings"
	"testing"
)

func TestBuiltinAll(t *testing.T) {
	a := Builtin()
	for _, name := range []string{"functional-logic", "security", "code-quality", "performance"} {
		if _, ok := a[name]; !ok {
			t.Errorf("missing builtin aspect: %q", name)
		}
	}
}

func TestLookupGeneric(t *testing.T) {
	a := Lookup("concurrency-safety")
	if a.Name != "concurrency-safety" {
		t.Errorf("name not propagated: %q", a.Name)
	}
	if len(a.ForbiddenKeywords) != 0 {
		t.Errorf("generic aspect should have no forbidden keywords, got %v", a.ForbiddenKeywords)
	}
	if !strings.Contains(a.SystemPrompt, "concurrency-safety") {
		t.Error("generic prompt should embed the aspect name")
	}
}

func TestAssemble(t *testing.T) {
	a := Lookup("security")
	got := Assemble(a, 1, 3, "Prior rounds: r1, r2")
	if !strings.Contains(got, "Round: 3 (critic-1)") {
		t.Errorf("missing round marker: %q", got[:200])
	}
	if !strings.Contains(got, "Prior rounds: r1, r2") {
		t.Error("missing prior rounds note")
	}
}
