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

func TestGoToTurnsWithoutRewind(t *testing.T) {
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
	s.R().Progress = 0.6
	s.R().State = RoverMoving
	held := s.R().Path[0]
	if err := s.GoTo(b.ID()); err != nil {
		t.Fatal(err)
	}
	if s.R().Progress < 0.59 {
		t.Fatalf("progress reset to %v — position was lost", s.R().Progress)
	}
	if s.R().Path[0] != held {
		t.Fatalf("current edge changed from %v to %v — rover would rewind", held, s.R().Path[0])
	}
	if s.R().Reversing {
		t.Fatal("retarget should not reverse back to the previous hex")
	}
	end := s.R().Path[len(s.R().Path)-1]
	if end != b {
		t.Fatalf("path end=%v want %v", end, b)
	}
	if s.R().Q != from.Q || s.R().R != from.R {
		t.Fatal("committed hex should stay until the edge finishes")
	}
}

func TestRetargetKeepsMoving(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.Status = StatusActive
	from := Axial{Q: s.R().Q, R: s.R().R}
	var a, b Axial
	n := 0
	for _, h := range s.Map {
		if h.Axial() == from || h.Impassable {
			continue
		}
		if CubeDistance(from, h.Axial()) != 1 {
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
		t.Skip("need two neighbors")
	}
	_ = s.GoTo(a.ID())
	s.R().Progress = 0.6
	s.R().State = RoverMoving
	held := s.R().Path[0]
	_ = s.GoTo(b.ID())
	if s.R().Reversing {
		t.Fatal("should keep the current edge, not back up")
	}
	if s.R().Path[0] != held {
		t.Fatalf("must finish current hex first, path[0]=%v held=%v", s.R().Path[0], held)
	}
	s.Tick(TickDT)
	if s.R().Progress <= 0.6 {
		t.Fatalf("should keep advancing along the same edge, progress=%v", s.R().Progress)
	}
	if s.R().Q != from.Q || s.R().R != from.R {
		t.Fatal("must stay on the committed hex until the edge completes")
	}
}

func TestQueuedLostWhenPickupDark(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.Status = StatusActive
	var c *Contract
	for i := range s.Contracts {
		if s.Contracts[i].Status == ContractQueued {
			c = &s.Contracts[i]
			break
		}
	}
	if c == nil {
		t.Fatal("need queued contract")
	}
	ax, err := ParseHexID(c.Pickup)
	if err != nil {
		t.Fatal(err)
	}
	s.Terminator.Direction = DirEast
	s.Terminator.Pos = float64(ax.Q) + 2
	if !s.InShadow(ax.Q, ax.R) {
		t.Fatal("pickup should be in shadow")
	}
	s.Tick(TickDT)
	if c.Status != ContractLostShadow {
		t.Fatalf("status=%s want lost_to_shadow", c.Status)
	}
}

func TestShiftContinuesAfterJobsFail(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.Status = StatusActive
	for i := range s.Contracts {
		s.Contracts[i].Status = ContractExpired
	}
	s.Tick(TickDT)
	if s.Status != StatusActive {
		t.Fatalf("status=%s — shift must not end just because jobs expired", s.Status)
	}
}

func TestSeedHasImpossibleAndHeavy(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	heavy, impossible := false, false
	for _, c := range s.Contracts {
		if c.Weight == WeightClassHeavy {
			heavy = true
		}
		if c.Impossible {
			impossible = true
		}
	}
	if !heavy {
		t.Fatal("need a heavy contract")
	}
	if !impossible {
		t.Fatal("need one intentionally near-impossible delivery")
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

func TestContinueJobAfterPickup(t *testing.T) {
	s := NewGame("MCC-TEST", RoverHauler)
	var c *Contract
	for i := range s.Contracts {
		if s.Contracts[i].Weight != WeightClassHeavy {
			c = &s.Contracts[i]
			break
		}
	}
	if c == nil {
		t.Fatal("need a contract")
	}
	if err := s.Accept(c.ID); err != nil {
		t.Fatal(err)
	}
	ax, err := ParseHexID(c.Pickup)
	if err != nil {
		t.Fatal(err)
	}
	s.R().Q, s.R().R = ax.Q, ax.R
	s.Status = StatusActive
	s.tryPickupDrop()
	s.continueJob()
	if c.Pickup == c.Dropoff {
		return
	}
	if len(s.R().Path) == 0 {
		t.Fatal("should auto-route to dropoff")
	}
	if s.R().Path[len(s.R().Path)-1].ID() != c.Dropoff {
		t.Fatalf("end=%v drop=%s", s.R().Path[len(s.R().Path)-1], c.Dropoff)
	}
}
