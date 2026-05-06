package ledger

import (
	"testing"
	"time"

	"latere.ai/x/debate/internal/state"
)

func TestPendingDeterministicOrder(t *testing.T) {
	agg := map[string]Record{
		"c1-3": {AttackID: "c1-3", Status: StatusOpen},
		"c1-1": {AttackID: "c1-1", Status: StatusOpen},
		"c1-2": {AttackID: "c1-2", Status: StatusRebutted},
		"c1-4": {AttackID: "c1-4", Status: StatusConceded},
	}
	got := Pending(agg)
	if len(got) != 3 {
		t.Fatalf("got %d, want 3", len(got))
	}
	for i, want := range []string{"c1-1", "c1-2", "c1-3"} {
		if got[i].AttackID != want {
			t.Errorf("index %d: got %q, want %q", i, got[i].AttackID, want)
		}
	}
}

func TestAggregateMissingFile(t *testing.T) {
	sess, err := state.NewSession(t.TempDir(), 1, ts())
	if err != nil {
		t.Fatal(err)
	}
	agg, err := Aggregate(sess)
	if err != nil {
		t.Fatal(err)
	}
	if len(agg) != 0 {
		t.Errorf("missing attacks.jsonl should yield empty map, got %d", len(agg))
	}
}

func TestLoadBodyNoSpill(t *testing.T) {
	sess, err := state.NewSession(t.TempDir(), 1, ts())
	if err != nil {
		t.Fatal(err)
	}
	// No spill: LoadBody is a no-op.
	r := Record{AttackID: "c1-1", Claim: "x"}
	got, err := LoadBody(sess, r)
	if err != nil {
		t.Fatal(err)
	}
	if got.Claim != "x" {
		t.Errorf("got %q", got.Claim)
	}
}

func ts() time.Time { return time.Now() }
