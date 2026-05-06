package round

import "testing"

func TestSteadyStateRequiresThree(t *testing.T) {
	d := &Detector{MaxRounds: 6, CostCap: 50000}
	if d.SteadyState([]ForkHistory{{NewAttacks: 0}, {NewAttacks: 0}}) {
		t.Error("two rounds should not trigger steady state")
	}
	hist := []ForkHistory{{NewAttacks: 4}, {NewAttacks: 0}, {NewAttacks: 0}}
	if !d.SteadyState(hist) {
		t.Error("expected steady state after two zero rounds")
	}
}

func TestSteadyStateResetsOnReattack(t *testing.T) {
	d := &Detector{}
	hist := []ForkHistory{{NewAttacks: 0}, {NewAttacks: 0, ReAttacks: 1}, {NewAttacks: 0, ReAttacks: 0}}
	if d.SteadyState(hist) {
		t.Error("re-attacks should suppress steady-state")
	}
}

func TestMalformedTwice(t *testing.T) {
	d := &Detector{}
	if !d.MalformedTwice([]ForkHistory{{MalformedFlag: true}, {MalformedFlag: true}}) {
		t.Error("two malformed should trigger")
	}
	if d.MalformedTwice([]ForkHistory{{MalformedFlag: false}, {MalformedFlag: true}}) {
		t.Error("one good resets malformed-twice")
	}
}

func TestCostMeter(t *testing.T) {
	m := NewCostMeter(50000)
	for _, n := range []int{1000, 5000, 10000, 40000} {
		m.Add(n)
	}
	if !m.ExceedsCap() {
		t.Errorf("cap should be exceeded; used=%d", m.Used())
	}
}

func TestEstimateTokens(t *testing.T) {
	if got := EstimateTokens("hello"); got != 2 {
		t.Errorf("got %d, want 2 (5/4 rounded up)", got)
	}
}

func TestMaxRoundsReached(t *testing.T) {
	d := &Detector{MaxRounds: 6}
	if d.MaxRoundsReached(5) {
		t.Error("round 5 should not reach max rounds 6")
	}
	if !d.MaxRoundsReached(6) {
		t.Error("round 6 should reach max rounds 6")
	}
	if !d.MaxRoundsReached(99) {
		t.Error("round 99 should reach max rounds 6")
	}
}

func TestCostCapHit(t *testing.T) {
	d := &Detector{CostCap: 1000}
	if d.CostCapHit(999) {
		t.Error("999 should not hit cap 1000")
	}
	if !d.CostCapHit(1000) {
		t.Error("1000 should hit cap 1000")
	}
	if !d.CostCapHit(1500) {
		t.Error("1500 should hit cap 1000")
	}
}

func TestCostMeterRemaining(t *testing.T) {
	m := NewCostMeter(100)
	if got := m.Remaining(); got != 100 {
		t.Errorf("fresh meter: got %d, want 100", got)
	}
	m.Add(40)
	if got := m.Remaining(); got != 60 {
		t.Errorf("after 40 used: got %d, want 60", got)
	}
	m.Add(80)
	if got := m.Remaining(); got != -20 {
		t.Errorf("after over-spend: got %d, want -20", got)
	}
}
