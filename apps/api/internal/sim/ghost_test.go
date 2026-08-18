package sim

import "testing"

func TestBuildGhostReplay(t *testing.T) {
	st := NewGame("MCC-TEST", RoverSwift)
	st.Status = StatusActive
	st.emit("hex_entered", map[string]any{"hexId": "0,1"})
	st.emit("hex_entered", map[string]any{"hexId": "1,1"})
	gr := BuildGhostReplay(st)
	if len(gr.Points) != 2 {
		t.Fatalf("points=%d", len(gr.Points))
	}
}

func TestVerdictDeterministic(t *testing.T) {
	a := VerdictText(OutcomeSaved, "MCC-7F2A")
	b := VerdictText(OutcomeSaved, "MCC-7F2A")
	if a != b {
		t.Fatal("verdict not deterministic")
	}
	if VerdictText(OutcomeLost, "MCC-7F2A") == a {
		t.Fatal("different outcomes should differ")
	}
}
