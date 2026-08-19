package sim

import (
	"math"
	"testing"
)

func TestTerminatorMovement(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.Status = StatusActive
	s.Terminator.Speed = 0.1
	s.Terminator.Direction = DirEast
	start := s.Terminator.Pos
	s.Tick(1.0)
	if math.Abs(s.Terminator.Pos-(start+0.1)) > 1e-9 {
		t.Fatalf("pos=%v want %v", s.Terminator.Pos, start+0.1)
	}
}

func TestNoShadowAtStart(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	for _, h := range s.Map {
		if s.Terminator.InShadow(h.Q) {
			t.Fatalf("hex %s in shadow at t=0 pos=%v", h.ID(), s.Terminator.Pos)
		}
	}
}

func TestShadowDrainsBattery(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.Status = StatusActive
	s.R().Battery = 50
	s.Terminator.Direction = DirEast
	s.Terminator.Pos = float64(s.R().Q) + 1 // center already passed
	if !s.InShadow(s.R().Q, s.R().R) {
		t.Fatal("expected rover in shadow")
	}
	s.Tick(1.0)
	if s.R().Battery >= 50 {
		t.Fatalf("battery=%v, should drain", s.R().Battery)
	}
}

func TestMoveCostRidgeGreaterThanRegolith(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	reg := Hex{Q: 1, R: 0, Type: TypeRegolith}
	ridge := Hex{Q: 1, R: 0, Type: TypeRidge}
	if s.MoveCost(ridge) <= s.MoveCost(reg) {
		t.Fatalf("ridge=%v reg=%v", s.MoveCost(ridge), s.MoveCost(reg))
	}
}

func TestPathfindingAvoidsImpassable(t *testing.T) {
	s := &GameState{
		Status: StatusActive,
		Map: map[string]Hex{
			"0,0":  {Q: 0, R: 0, Type: TypeRegolith},
			"1,0":  {Q: 1, R: 0, Type: TypeCrater, Impassable: true},
			"2,0":  {Q: 2, R: 0, Type: TypeRegolith},
			"0,-1": {Q: 0, R: -1, Type: TypeRegolith},
			"1,-1": {Q: 1, R: -1, Type: TypeRegolith},
			"2,-1": {Q: 2, R: -1, Type: TypeRegolith},
		},
		Rovers:     []Rover{NewRover(RoverSwift, 0, 0)},
		Terminator: Terminator{Pos: -10, Speed: TerminatorSpeed, Direction: DirEast},
	}
	path := s.FindPath(Axial{0, 0}, Axial{2, 0})
	if len(path) == 0 {
		t.Fatal("expected path")
	}
	for _, a := range path {
		if a.ID() == "1,0" {
			t.Fatal("path went through impassable crater")
		}
	}
	if path[len(path)-1] != (Axial{2, 0}) {
		t.Fatalf("end=%v", path[len(path)-1])
	}
}

func TestSeedDeterministic(t *testing.T) {
	a := NewGame("MCC-7F2A", RoverSwift)
	b := NewGame("MCC-7F2A", RoverSwift)
	if a.CrisisKind != b.CrisisKind || a.CrisisAt != b.CrisisAt {
		t.Fatal("crisis mismatch")
	}
	if a.Terminator.Direction != b.Terminator.Direction {
		t.Fatal("dir mismatch")
	}
	if len(a.Map) != len(b.Map) || len(a.Map) != MapHexCount {
		t.Fatalf("map size %d vs %d", len(a.Map), len(b.Map))
	}
	for id, h := range a.Map {
		o, ok := b.Map[id]
		if !ok || o != h {
			t.Fatalf("hex %s mismatch %#v %#v", id, h, o)
		}
	}
	if len(a.Contracts) != len(b.Contracts) {
		t.Fatal("contracts len")
	}
	for i := range a.Contracts {
		if a.Contracts[i] != b.Contracts[i] {
			t.Fatalf("contract %d mismatch", i)
		}
	}
	if len(a.Salvage) != len(b.Salvage) {
		t.Fatal("salvage len")
	}
	for i := range a.Salvage {
		if a.Salvage[i] != b.Salvage[i] {
			t.Fatalf("salvage %d mismatch", i)
		}
	}
}

func TestDifferentSeedsDiffer(t *testing.T) {
	a := NewGame("MCC-AAAA", RoverSwift)
	b := NewGame("MCC-BBBB", RoverSwift)
	if a.CrisisKind == b.CrisisKind && a.Terminator.Direction == b.Terminator.Direction {
		same := true
		for id, h := range a.Map {
			if b.Map[id] != h {
				same = false
				break
			}
		}
		if same {
			t.Fatal("expected different layouts or crises")
		}
	}
}

func TestBatteryZeroStrandsWhenMoving(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.Status = StatusActive
	s.R().Battery = 0.01
	s.R().State = RoverMoving
	from := Axial{Q: s.R().Q, R: s.R().R}
	var path []Axial
	for _, d := range Neighbors {
		nb := Axial{Q: from.Q + d.Q, R: from.R + d.R}
		if h, ok := s.Map[nb.ID()]; ok && !h.Impassable {
			path = []Axial{nb}
			break
		}
	}
	if len(path) == 0 {
		t.Fatal("need a neighbor path")
	}
	s.R().Path = path
	for i := 0; i < 80; i++ {
		s.Tick(TickDT)
		if s.R().State == RoverStranded {
			return
		}
	}
	t.Fatalf("state=%s battery=%v", s.R().State, s.R().Battery)
}

func TestPredictMatchesMoveCost(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	from := Axial{Q: s.R().Q, R: s.R().R}
	var to Axial
	found := false
	for _, d := range Neighbors {
		nb := Axial{Q: from.Q + d.Q, R: from.R + d.R}
		if _, ok := s.Map[nb.ID()]; ok {
			to = nb
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no neighbor")
	}
	path := []Axial{to}
	pred := s.Predict(path)
	if pred.ETASec <= 0 {
		t.Fatalf("eta %v", pred.ETASec)
	}
	hex := s.Map[to.ID()]
	cost := s.MoveCost(hex)
	if pred.EndBattery >= s.R().Battery {
		t.Fatalf("expected drain, start=%v end=%v cost=%v", s.R().Battery, pred.EndBattery, cost)
	}
}

func TestGoToFindsPath(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	from := Axial{Q: s.R().Q, R: s.R().R}
	var to Axial
	okDest := false
	for _, h := range s.Map {
		if h.Axial() == from || h.Impassable {
			continue
		}
		to = h.Axial()
		okDest = true
		break
	}
	if !okDest {
		t.Fatal("no dest")
	}
	if err := s.GoTo(to.ID()); err != nil {
		t.Fatal(err)
	}
	if len(s.R().Path) == 0 {
		t.Fatal("empty path")
	}
	if s.R().Path[len(s.R().Path)-1] != to {
		t.Fatalf("end=%v want %v", s.R().Path[len(s.R().Path)-1], to)
	}
}
