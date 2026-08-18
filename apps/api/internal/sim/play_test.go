package sim

import "testing"

func TestSwiftRejectsHeavy(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	id := ""
	for _, c := range s.Contracts {
		if c.Weight == WeightClassHeavy {
			id = c.ID
			break
		}
	}
	if id == "" {
		t.Skip("no heavy contract on this seed")
	}
	if err := s.Accept(id); err == nil {
		t.Fatal("swift should reject heavy")
	}
}

func TestHaulerAcceptsHeavy(t *testing.T) {
	s := NewGame("MCC-TEST", RoverHauler)
	id := ""
	for _, c := range s.Contracts {
		if c.Weight == WeightClassHeavy {
			id = c.ID
			break
		}
	}
	if id == "" {
		t.Skip("no heavy contract on this seed")
	}
	if err := s.Accept(id); err != nil {
		t.Fatal(err)
	}
}

func TestDispatchStartsMoving(t *testing.T) {
	s := NewGame("MCC-TEST", RoverHauler)
	var id string
	for _, c := range s.Contracts {
		if c.Weight != WeightClassHeavy {
			id = c.ID
			break
		}
	}
	if id == "" {
		t.Fatal("need a contract")
	}
	if err := s.Dispatch(id); err != nil {
		t.Fatal(err)
	}
	if s.R().State != RoverMoving && len(s.R().Path) == 0 {
		t.Fatalf("state=%s path=%d", s.R().State, len(s.R().Path))
	}
}

func TestLowBatteryDoesNotDeploy(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.R().Battery = 1
	from := Axial{Q: s.R().Q, R: s.R().R}
	var to Axial
	for _, h := range s.Map {
		if h.Axial() == from || h.Impassable {
			continue
		}
		if CubeDistance(from, h.Axial()) >= 2 {
			to = h.Axial()
			break
		}
	}
	if to == (Axial{}) {
		t.Skip("no far hex")
	}
	if err := s.GoTo(to.ID()); err != nil {
		t.Fatal(err)
	}
	if s.R().State == RoverMoving {
		t.Fatal("should not deploy an infeasible route")
	}
	if len(s.R().Path) == 0 {
		t.Fatal("path should still be shown")
	}
}

func TestGoToKeepsProgress(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	from := Axial{Q: s.R().Q, R: s.R().R}
	var a, b Axial
	n := 0
	for _, h := range s.Map {
		if h.Axial() == from || h.Impassable {
			continue
		}
		if n == 0 {
			a = h.Axial()
		} else {
			b = h.Axial()
			break
		}
		n++
	}
	if a == (Axial{}) || b == (Axial{}) {
		t.Fatal("need two dests")
	}
	if err := s.GoTo(a.ID()); err != nil {
		t.Fatal(err)
	}
	if len(s.R().Path) == 0 {
		t.Fatal("no path")
	}
	next := s.R().Path[0]
	s.R().Progress = 0.4
	s.R().State = RoverMoving
	if err := s.GoTo(b.ID()); err != nil {
		t.Fatal(err)
	}
	if s.R().Progress < 0.39 {
		t.Fatalf("progress reset to %v", s.R().Progress)
	}
	if len(s.R().Path) == 0 || s.R().Path[0] != next {
		t.Fatalf("current edge dropped: %v want %v", s.R().Path, next)
	}
}

func TestTwoRoversSpawn(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	if len(s.Rovers) != 2 {
		t.Fatalf("rovers=%d", len(s.Rovers))
	}
	s.SelectRover(RoverHauler)
	if s.R().Type != RoverHauler {
		t.Fatal("select hauler")
	}
}

func TestTwoRoversMoveIndependently(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	var light string
	for _, c := range s.Contracts {
		if c.Weight != WeightClassHeavy {
			light = c.ID
			break
		}
	}
	if light == "" {
		t.Fatal("need light/medium")
	}
	if err := s.Dispatch(light); err != nil {
		t.Fatal(err)
	}
	swiftPath := len(s.R().Path)
	s.SelectRover(RoverHauler)
	var heavy string
	for _, c := range s.Contracts {
		if c.Weight == WeightClassHeavy && c.Status == ContractQueued {
			heavy = c.ID
			break
		}
	}
	if heavy == "" {
		t.Skip("no free heavy")
	}
	if err := s.Dispatch(heavy); err != nil {
		t.Fatal(err)
	}
	if s.R().Type != RoverHauler {
		t.Fatal("hauler should be selected")
	}
	if s.Rovers[0].State != RoverMoving {
		t.Fatalf("swift should keep moving, state=%s path=%d", s.Rovers[0].State, swiftPath)
	}
	if s.Rovers[1].State != RoverMoving && len(s.Rovers[1].Path) == 0 {
		t.Fatal("hauler should have a job")
	}
}

func TestContractsHaveRiskAndReward(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	if len(s.Contracts) == 0 {
		t.Fatal("no contracts")
	}
	for _, c := range s.Contracts {
		if c.Reward != c.ColonyValue+c.EarthValue {
			t.Fatalf("reward %d", c.Reward)
		}
		if c.Risk == "" || c.Urgency == "" {
			t.Fatalf("missing risk/urgency %#v", c)
		}
	}
}
