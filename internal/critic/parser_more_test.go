package critic

import (
	"strings"
	"testing"
)

func TestWithdrawValidAndInvalid(t *testing.T) {
	docInvalid := "# Critic 1 — round 3 attacks\n\naspect: security\n\n## c1-9 [x:1] (withdraw)\n\nreason: not real\n"
	out, stats, err := Parse(docInvalid, "security", 1, 3, nil, ParseOption{})
	if err != nil {
		t.Fatal(err)
	}
	// Unknown id under (withdraw) is treated as introduce; will be
	// dropped because there's no reproduction.
	if stats.DroppedNoReproduce == 0 {
		t.Errorf("withdraw of unknown id should fall back to introduce and drop without repro; got %+v", stats)
	}
	_ = out

	docValid := "# Critic 1 — round 3 attacks\n\naspect: security\n\n## c1-2 [x:1] (withdraw)\n\nreason: false positive\n"
	out, stats, err = Parse(docValid, "security", 1, 3, []string{"c1-2"}, ParseOption{})
	if err != nil {
		t.Fatal(err)
	}
	if stats.KeptWithdraw != 1 || len(out) != 1 {
		t.Fatalf("expected 1 kept withdraw, got %+v out=%d", stats, len(out))
	}
	if out[0].Disposition != DispWithdraw || out[0].WithdrawReason != "false positive" {
		t.Errorf("withdraw not preserved: %+v", out[0])
	}
}

func TestReAttackUnknownIdFallsBack(t *testing.T) {
	doc := "# Critic 1 — round 3 attacks\n\naspect: security\n\n" +
		"## c1-99 [x:1] (re-attack)\n\nclaim: tighter\n\nexpected violation: panic\n\nreproduction:\n```\ngo\n```\n"
	out, stats, err := Parse(doc, "security", 1, 3, []string{"c1-1"}, ParseOption{})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Disposition != DispIntroduce {
		t.Errorf("re-attack of unknown id should fall back to introduce: %+v", out)
	}
	if stats.Renamed == 0 {
		t.Error("expected renamed counter to fire")
	}
}

func TestRenderRoundTripWithWithdraw(t *testing.T) {
	atks := []Attack{
		{
			AttackID: "c1-1", Aspect: "security", Round: 1, Disposition: DispIntroduce,
			Location: "x:1", Claim: "leak", ExpectedViolation: "panic", Reproduction: "go test",
		},
		{
			AttackID: "c1-2", Aspect: "security", Round: 3, Disposition: DispWithdraw,
			Location: "y:1", WithdrawReason: "false positive",
		},
	}
	r := Render(1, 3, "security", atks)
	got := string(r)
	if !strings.Contains(got, "(withdraw)") {
		t.Errorf("missing withdraw tag: %s", got)
	}
	if !strings.Contains(got, "reason: false positive") {
		t.Error("missing reason line")
	}
}

func TestStyleAttackKeptWhenAllowed(t *testing.T) {
	doc := "# Critic 1 — round 1 attacks\n\naspect: code-quality\n\n## c1-1 [x:1]\n\nclaim: This function should be named more idiomatic.\n\nexpected violation: it bothers me\n\nreproduction:\n```\nrun\n```\n"
	out, stats, err := Parse(doc, "code-quality", 1, 1, nil, ParseOption{AllowStyleAttacks: true})
	if err != nil {
		t.Fatal(err)
	}
	if stats.DroppedStyle != 0 || len(out) != 1 {
		t.Errorf("style should be kept under AllowStyleAttacks: %+v", stats)
	}
}

func TestEmptyHeaderError(t *testing.T) {
	_, _, err := Parse("", "security", 1, 1, nil, ParseOption{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}
