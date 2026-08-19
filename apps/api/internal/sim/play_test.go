package sim

import (
	"fmt"
	"testing"
)

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
			n++
			continue
		}
		b = h.Axial()
		n++
		break
	}
	if n < 2 {
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

func TestSeedHasHeavy(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	heavy := false
	for _, c := range s.Contracts {
		if c.Weight == WeightClassHeavy {
			heavy = true
		}
		if c.Impossible {
			t.Fatal("no bait jobs — every contract should be a real delivery")
		}
	}
	if !heavy {
		t.Fatal("need a heavy contract")
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

func TestSwitchBackToSwiftCommandsSwift(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.SelectRover(RoverHauler)
	s.SelectRover(RoverSwift)
	if s.R().Type != RoverSwift {
		t.Fatal("swift should be selected")
	}
	from := Axial{Q: s.R().Q, R: s.R().R}
	var to Axial
	found := false
	for _, d := range Neighbors {
		nb := Axial{Q: from.Q + d.Q, R: from.R + d.R}
		if h, ok := s.Map[nb.ID()]; ok && !h.Impassable {
			to, found = nb, true
			break
		}
	}
	if !found {
		t.Fatal("need neighbor")
	}
	if err := s.GoTo(to.ID()); err != nil {
		t.Fatal(err)
	}
	s.Tick(TickDT)
	if s.Rovers[0].Type != RoverSwift {
		t.Fatal("index 0 is swift")
	}
	swiftHere := HexID(s.Rovers[0].Q, s.Rovers[0].R)
	if swiftHere != to.ID() && s.Rovers[0].State != RoverMoving && len(s.Rovers[0].Path) == 0 {
		t.Fatal("swift must take the goto")
	}
	if s.Rovers[1].State == RoverMoving || len(s.Rovers[1].Path) > 0 {
		t.Fatal("hauler moved after switching back to swift")
	}
}

func TestDispatchDoesNotStealSelection(t *testing.T) {
	s := NewGame("MCC-TEST", RoverHauler)
	var heavy string
	for _, c := range s.Contracts {
		if c.Weight == WeightClassHeavy && c.Status == ContractQueued {
			heavy = c.ID
			break
		}
	}
	if heavy == "" {
		t.Skip("no heavy")
	}
	if err := s.Dispatch(heavy); err != nil {
		t.Fatal(err)
	}
	s.SelectRover(RoverSwift)
	_ = s.Dispatch(heavy)
	if s.R().Type != RoverSwift {
		t.Fatal("clicking hauler's job must not silently switch back to hauler")
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

func TestDropoffsSitOnFarBase(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	want := s.farBaseID()
	for _, c := range s.Contracts {
		if c.Dropoff != want {
			t.Fatalf("dropoff %s want far base %s", c.Dropoff, want)
		}
		ax, err := ParseHexID(c.Dropoff)
		if err != nil {
			t.Fatal(err)
		}
		if s.Terminator.InShadow(ax.Q) {
			t.Fatal("dropoff starts in shadow")
		}
	}
}

func TestShadowedDropoffMovesToLitBase(t *testing.T) {
	s := NewGame("MCC-TEST", RoverHauler)
	s.Status = StatusActive
	var idx int
	for i := range s.Contracts {
		if s.Contracts[i].Weight == WeightClassHeavy {
			idx = i
			break
		}
	}
	c := &s.Contracts[idx]
	s.SelectRover(RoverHauler)
	far, err := ParseHexID(s.farBaseID())
	if err != nil {
		t.Fatal(err)
	}
	s.R().Q, s.R().R = far.Q, far.R
	c.Status = ContractInTransit
	c.AssignedTo = RoverHauler
	near := HexID(s.R().Q, s.R().R)
	if s.Terminator.Direction == DirEast {
		c.Dropoff = HexID(0, 0)
		s.Terminator.Pos = 1
	} else {
		c.Dropoff = HexID(MapCols-1, 0)
		s.Terminator.Pos = float64(MapCols) - 2
	}
	if !s.Terminator.InShadow(mustQ(c.Dropoff)) {
		t.Fatal("setup: dropoff should be dark")
	}
	s.Tick(TickDT)
	c = &s.Contracts[idx]
	if c.Dropoff == near && s.Terminator.InShadow(mustQ(c.Dropoff)) {
		t.Fatal("dropoff stayed in shadow")
	}
	if s.Terminator.InShadow(mustQ(c.Dropoff)) {
		t.Fatalf("dropoff %s still dark", c.Dropoff)
	}
}

func mustQ(id string) int {
	ax, err := ParseHexID(id)
	if err != nil {
		panic(err)
	}
	return ax.Q
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

func TestPickupsDoNotShareHex(t *testing.T) {
	for _, seed := range []string{"MCC-TEST", "MCC-7F2A", "MCC-AAAA", "MCC-BBBB", "MCC-H6SZ"} {
		s := NewGame(seed, RoverSwift)
		seen := map[string]string{}
		for _, c := range s.Contracts {
			if prev, ok := seen[c.Pickup]; ok {
				t.Fatalf("seed %s: pickup %s used by %s and %s", seed, c.Pickup, prev, c.ID)
			}
			seen[c.Pickup] = c.ID
		}
		for _, sv := range s.Salvage {
			if prev, ok := seen[sv.Hex]; ok {
				t.Fatalf("seed %s: salvage %s overlaps %s", seed, sv.Hex, prev)
			}
			seen[sv.Hex] = sv.ID
		}
		if len(s.Salvage) != SalvageCount {
			t.Fatalf("seed %s: salvage %d want %d", seed, len(s.Salvage), SalvageCount)
		}
	}
}

func TestCannotPathIntoShadow(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	from := Axial{Q: s.R().Q, R: s.R().R}
	s.Terminator.Direction = DirEast
	s.Terminator.Pos = 10
	var dark Axial
	foundDark := false
	for _, h := range s.Map {
		if h.Axial() == from {
			continue
		}
		if s.Terminator.InShadow(h.Q) {
			dark = h.Axial()
			foundDark = true
			break
		}
	}
	if !foundDark {
		t.Fatal("need a dark hex")
	}
	if path := s.FindPath(from, dark); len(path) != 0 {
		t.Fatalf("path entered shadow: %v", path)
	}
	if err := s.GoTo(dark.ID()); err == nil && s.R().State == RoverMoving {
		t.Fatal("must not deploy into shadow")
	}
}

func TestShadowOvertakeStrands(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.Status = StatusActive
	s.Terminator.Direction = DirEast
	s.Terminator.Pos = float64(s.R().Q) + 1
	if !s.InShadow(s.R().Q, s.R().R) {
		t.Fatal("rover should be in shadow")
	}
	s.Tick(TickDT)
	if s.R().State != RoverStranded {
		t.Fatalf("state=%s, shadow must strand", s.R().State)
	}
}

func greedyColony(seed string) int {
	s := NewGame(seed, RoverSwift)
	s.Status = StatusActive
	for s.Status == StatusActive {
		sel := s.Active
		for i := range s.Rovers {
			s.Active = i
			r := s.R()
			if r.State != RoverIdle {
				continue
			}
			best := -1
			bestScore := -1
			for j, c := range s.Contracts {
				if c.Status != ContractQueued {
					continue
				}
				if reason := s.acceptBlocked(c); reason != "" {
					continue
				}
				if c.ColonyValue > bestScore {
					bestScore = c.ColonyValue
					best = j
				}
			}
			if best >= 0 {
				_ = s.Dispatch(s.Contracts[best].ID)
			} else {
				s.seekSalvage()
			}
		}
		s.Active = sel
		s.Tick(TickDT)
	}
	return s.ColonyScore
}

func TestDeliverableCeilingAllowsWin(t *testing.T) {
	for i := 0; i < 16; i++ {
		s := NewGame(fmt.Sprintf("MCC-C%02d", i), RoverSwift)
		if deliverableCeiling(s) < ColonyWinThreshold {
			t.Fatalf("seed MCC-C%02d ceiling %d < %d", i, deliverableCeiling(s), ColonyWinThreshold)
		}
	}
}

func TestSolverCanReachColonyWin(t *testing.T) {
	wins, max := 0, 0
	n := 12
	for i := 0; i < n; i++ {
		score := densityColony(fmt.Sprintf("MCC-S%02d", i))
		if score > max {
			max = score
		}
		if score >= ColonyWinThreshold {
			wins++
		}
	}
	t.Logf("density solver wins %d/%d max=%d need=%d", wins, n, max, ColonyWinThreshold)
	if max < ColonyWinThreshold {
		t.Fatalf("no seed reached 100 (max=%d) — 100+ must be possible", max)
	}
	if wins < 2 {
		t.Fatalf("solver won %d/%d — 100 should be reachable on more than a fluke", wins, n)
	}
}

func TestDeepSearchBeatsThresholdOnASeed(t *testing.T) {
	best := 0
	for _, seed := range []string{"MCC-S00", "MCC-S03", "MCC-S07"} {
		sc := searchColony(seed)
		if sc > best {
			best = sc
		}
		if sc >= ColonyWinThreshold {
			t.Logf("%s search colony=%d", seed, sc)
			return
		}
	}
	t.Fatalf("deep assignment search never reached 100 (best=%d)", best)
}

func TestGreedyWinRateIsAChallenge(t *testing.T) {
	wins := 0
	n := 20
	sum := 0
	for i := 0; i < n; i++ {
		score := greedyColony(fmt.Sprintf("MCC-G%02d", i))
		sum += score
		if score >= ColonyWinThreshold {
			wins++
		}
	}
	avg := float64(sum) / float64(n)
	t.Logf("greedy wins %d/%d avg colony=%.1f need=%d", wins, n, avg, ColonyWinThreshold)
	if wins == n {
		t.Fatalf("greedy won every seed — too easy")
	}
	if wins >= n-1 {
		t.Fatalf("greedy won %d/%d — 100 should still require both rovers and triage", wins, n)
	}
	if wins == 0 && avg < float64(ColonyWinThreshold)/4 {
		t.Fatalf("greedy never scores — too hard (avg=%.1f)", avg)
	}
}

func TestSalvagePickupAddsColony(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	if len(s.Salvage) == 0 {
		t.Fatal("expected salvage caches")
	}
	s.Status = StatusActive
	sv := s.Salvage[0]
	ax, err := ParseHexID(sv.Hex)
	if err != nil {
		t.Fatal(err)
	}
	s.R().Q, s.R().R = ax.Q, ax.R
	s.R().State = RoverIdle
	before := s.ColonyScore
	s.trySalvage()
	if s.ColonyScore != before+sv.Value {
		t.Fatalf("score=%d want %d", s.ColonyScore, before+sv.Value)
	}
	if s.Salvage[0].Status != SalvageTaken {
		t.Fatalf("status=%s", s.Salvage[0].Status)
	}
}

func TestSalvageLostToShadow(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.Status = StatusActive
	sv := s.Salvage[0]
	ax, _ := ParseHexID(sv.Hex)
	s.Terminator.Direction = DirEast
	s.Terminator.Pos = float64(ax.Q) + 1
	s.tickSalvage()
	if s.Salvage[0].Status != SalvageLost {
		t.Fatalf("status=%s", s.Salvage[0].Status)
	}
}

func TestShiftContinuesForSalvageAfterContracts(t *testing.T) {
	s := NewGame("MCC-TEST", RoverSwift)
	s.Status = StatusActive
	s.CrisisFired = true
	s.CrisisKind = "dust_storm"
	for i := range s.Contracts {
		s.Contracts[i].Status = ContractDelivered
	}
	s.ColonyScore = 90
	s.checkEnd()
	if s.Status != StatusActive {
		t.Fatalf("status=%s, salvage should keep the shift open", s.Status)
	}
	for i := range s.Salvage {
		s.Salvage[i].Status = SalvageTaken
	}
	s.checkEnd()
	if s.Status != StatusFinished || s.EndReason != "delivered" {
		t.Fatalf("status=%s reason=%s", s.Status, s.EndReason)
	}
}
