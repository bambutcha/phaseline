package sim

import (
	"math"
	"sort"
)

func (s *GameState) Tick(dt float64) []Event {
	if s.Status != StatusActive {
		return nil
	}
	start := len(s.Events)
	s.T += dt
	s.Terminator = s.Terminator.Advance(dt)

	sel := s.Active
	for i := range s.Rovers {
		s.Active = i
		s.tickRover(dt)
	}
	s.Active = sel

	s.tickCargo(dt)
	s.tickCrisis()
	s.checkEnd()
	return s.Events[start:]
}

func (s *GameState) tickRover(dt float64) {
	r := s.R()
	if r.State == RoverMoving {
		s.tickMove(dt)
	}
	s.tickBattery(dt)
	if r.Battery <= 0 {
		r.Battery = 0
		if r.State == RoverMoving {
			s.failCarried(ContractFailed)
			r.State = RoverStranded
			r.Path = nil
			s.emit("stranded", map[string]any{"rover": string(r.Type), "hexId": HexID(r.Q, r.R)})
		}
	}
	if r.Battery <= r.MaxBattery*0.2 && r.Battery > 0 {
		s.emitOnce("battery_low_"+string(r.Type), map[string]any{"rover": string(r.Type), "battery": r.Battery})
	}
}

func (s *GameState) tickMove(dt float64) {
	r := s.R()
	if r.Reversing {
		s.tickReverse(dt)
		return
	}
	if len(r.Path) == 0 {
		r.State = RoverIdle
		r.Progress = 0
		s.continueJob()
		return
	}
	next := r.Path[0]
	hex, ok := s.Map[next.ID()]
	if !ok || hex.Impassable {
		r.State = RoverIdle
		r.Path = nil
		r.Progress = 0
		return
	}
	cost := s.MoveCost(hex)
	if math.IsInf(cost, 1) {
		r.State = RoverIdle
		return
	}
	speed := s.roverSpeed()
	edgeTime := cost / speed
	if edgeTime <= 0 {
		edgeTime = TickDT
	}
	r.Progress += dt / edgeTime
	if r.Progress < 1 {
		return
	}

	prevShadow := s.InShadow(r.Q, r.R)
	r.Battery -= cost
	if r.Battery <= 0 {
		r.Battery = 0
		r.Q, r.R = next.Q, next.R
		r.Path = nil
		r.Progress = 0
		s.failCarried(ContractFailed)
		r.State = RoverStranded
		s.emit("stranded", map[string]any{"rover": string(r.Type), "hexId": HexID(r.Q, r.R)})
		return
	}
	r.Q, r.R = next.Q, next.R
	r.Path = r.Path[1:]
	r.Progress = 0
	r.SunIdle = 0
	if r.PanicLeft > 0 {
		r.PanicLeft--
	}
	s.emit("hex_entered", map[string]any{"rover": string(r.Type), "hexId": next.ID()})

	nowShadow := s.InShadow(r.Q, r.R)
	if nowShadow && !prevShadow {
		s.emit("entered_shadow", map[string]any{"hexId": next.ID()})
	}
	if hex.Type == TypeCrater || hex.Type == TypeRidge {
		if s.hasCargo(CargoCrew) {
			r.PanicLeft = 1
		}
	}
	s.tryPickupDrop()
	if len(r.Path) == 0 {
		r.State = RoverIdle
		s.continueJob()
	}
	if r.Battery <= 0 {
		r.Battery = 0
	}
}

func (s *GameState) tickReverse(dt float64) {
	r := s.R()
	hex, ok := s.Map[r.ReverseTo.ID()]
	if !ok {
		r.Reversing = false
		r.Progress = 0
		if len(r.Path) > 0 {
			r.State = RoverMoving
		} else {
			r.State = RoverIdle
			s.continueJob()
		}
		return
	}
	cost := s.MoveCost(hex)
	speed := s.roverSpeed()
	edgeTime := cost / speed
	if edgeTime <= 0 {
		edgeTime = TickDT
	}
	r.Progress -= dt / edgeTime
	if r.Progress > 0 {
		return
	}
	r.Progress = 0
	r.Reversing = false
	if len(r.Path) == 0 {
		r.State = RoverIdle
		s.continueJob()
		return
	}
	r.State = RoverMoving
}

func (s *GameState) continueJob() {
	if s.noChain {
		return
	}
	r := s.R()
	if r.State == RoverStranded || r.Reversing {
		return
	}
	if len(r.Path) > 0 {
		return
	}
	here := HexID(r.Q, r.R)
	var next string
	for _, c := range s.Contracts {
		if c.AssignedTo != r.Type {
			continue
		}
		if c.Status == ContractInTransit && c.Dropoff != here {
			next = c.Dropoff
			break
		}
	}
	if next == "" {
		for _, c := range s.Contracts {
			if c.AssignedTo != r.Type {
				continue
			}
			if c.Status == ContractAccepted && c.Pickup != here {
				next = c.Pickup
				break
			}
		}
	}
	if next == "" || next == here {
		return
	}
	_ = s.GoTo(next)
}

func (s *GameState) tickBattery(dt float64) {
	r := s.R()
	here := s.RoverHex()
	inShadow := s.InShadow(r.Q, r.R)
	drain := r.IdleDrain * dt
	if inShadow {
		drain *= 2
		if s.hasCargo(CargoO2) {
			drain *= 2
		}
	}
	r.Battery -= drain

	if !inShadow && here.Type == TypeSolarPlateau && !s.FlareActive {
		r.Battery += SolarGain * dt
	}
	if r.State == RoverIdle && !inShadow && here.Type == TypeBase {
		r.Battery += BaseRecharge * dt
	}
	if r.Battery > r.MaxBattery {
		r.Battery = r.MaxBattery
	}
	if r.State == RoverIdle && !inShadow {
		r.SunIdle += dt
	} else if inShadow {
		r.SunIdle = 0
	}
}

func (s *GameState) tickCargo(dt float64) {
	for i := range s.Contracts {
		c := &s.Contracts[i]
		if c.Status == ContractAccepted || c.Status == ContractInTransit {
			if c.Deadline > 0 {
				c.Deadline -= dt
				if c.Deadline <= 0 {
					c.Deadline = 0
					c.Status = ContractExpired
					s.dropCargoOf(c.AssignedTo, c.Cargo)
					s.emit("contract_failed", map[string]any{"contractId": c.ID, "reason": "expired"})
				}
			}
		}
		if c.Status == ContractInTransit && c.Cargo == CargoCryo {
			if rv := s.roverByType(c.AssignedTo); rv != nil && rv.SunIdle >= 1 {
				c.Status = ContractFailed
				s.dropCargoOf(c.AssignedTo, CargoCryo)
				s.emit("contract_failed", map[string]any{"contractId": c.ID, "reason": "spoil"})
			}
		}
	}
}

func (s *GameState) roverByType(t RoverType) *Rover {
	for i := range s.Rovers {
		if s.Rovers[i].Type == t {
			return &s.Rovers[i]
		}
	}
	return s.R()
}

func (s *GameState) tickCrisis() {
	if s.CrisisFired || s.Status != StatusActive {
		return
	}
	if s.T < s.CrisisAt {
		return
	}
	s.CrisisFired = true
	s.emit("crisis", map[string]any{"kind": s.CrisisKind})
	switch s.CrisisKind {
	case "dust_storm":
		s.DustStorm = true
	case "solar_flare":
		s.FlareActive = true
	case "cave_in":
		s.caveIn()
	case "comm_blackout":
		s.CommUntil = s.T + 15
	case "vip_override":
		s.addVIP()
	}
}

func (s *GameState) caveIn() {
	ids := make([]string, 0, len(s.Map))
	for id := range s.Map {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	busy := map[string]bool{}
	for _, r := range s.Rovers {
		busy[HexID(r.Q, r.R)] = true
	}
	for _, id := range ids {
		h := s.Map[id]
		if h.Type == TypeCrater && !h.Impassable && !busy[id] {
			h.Impassable = true
			s.Map[id] = h
			return
		}
	}
}

func (s *GameState) addVIP() {
	r := s.R()
	drop := HexID(r.Q, r.R)
	ids := make([]string, 0)
	for id, h := range s.Map {
		if h.Type == TypeBase {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	if len(ids) > 0 {
		drop = ids[0]
	}
	s.Contracts = append(s.Contracts, Contract{
		ID:          "vip",
		Title:       "VIP Override — срочный груз",
		Cargo:       CargoMedSeeds,
		Weight:      WeightClassLight,
		Pickup:      HexID(r.Q, r.R),
		Dropoff:     drop,
		ColonyValue: 40,
		EarthValue:  10,
		Reward:      50,
		Risk:        "high",
		Urgency:     "high",
		Deadline:    45,
		Status:      ContractQueued,
	})
}

func (s *GameState) tryPickupDrop() {
	r := s.R()
	here := HexID(r.Q, r.R)
	for i := range s.Contracts {
		c := &s.Contracts[i]
		if c.AssignedTo != "" && c.AssignedTo != r.Type {
			continue
		}
		if c.Status == ContractAccepted && c.Pickup == here {
			c.Status = ContractInTransit
			c.AssignedTo = r.Type
			r.Cargo = append(r.Cargo, c.Cargo)
			s.emit("pickup", map[string]any{"contractId": c.ID, "rover": string(r.Type)})
		}
		if c.Status == ContractInTransit && c.Dropoff == here {
			c.Status = ContractDelivered
			s.dropCargo(c.Cargo)
			bonus := 0
			if c.Cargo == CargoCommRelay && !s.CrisisFired {
				bonus = 15
			}
			s.ColonyScore += c.ColonyValue + bonus
			s.EarthScore += c.EarthValue
			s.emit("deliver", map[string]any{
				"contractId": c.ID, "hexId": here, "rover": string(r.Type),
				"colony": c.ColonyValue + bonus, "earth": c.EarthValue,
			})
		}
	}
}

func (s *GameState) hasCargo(t CargoType) bool {
	for _, c := range s.R().Cargo {
		if c == t {
			return true
		}
	}
	return false
}

func (s *GameState) dropCargo(t CargoType) {
	s.dropCargoOf(s.R().Type, t)
}

func (s *GameState) dropCargoOf(owner RoverType, t CargoType) {
	rv := s.roverByType(owner)
	if rv == nil {
		return
	}
	out := rv.Cargo[:0]
	for _, c := range rv.Cargo {
		if c != t {
			out = append(out, c)
		}
	}
	rv.Cargo = out
}

func (s *GameState) failCarried(st ContractStatus) {
	r := s.R()
	for i := range s.Contracts {
		c := &s.Contracts[i]
		if c.AssignedTo != "" && c.AssignedTo != r.Type {
			continue
		}
		if c.Status == ContractInTransit || c.Status == ContractAccepted {
			if st == ContractFailed && s.InShadow(r.Q, r.R) {
				c.Status = ContractLostShadow
			} else {
				c.Status = st
			}
			s.emit("contract_failed", map[string]any{"contractId": c.ID})
		}
	}
	r.Cargo = nil
}

func (s *GameState) checkEnd() {
	if s.Status != StatusActive {
		return
	}
	allDone := true
	for _, c := range s.Contracts {
		switch c.Status {
		case ContractQueued, ContractAccepted, ContractInTransit:
			allDone = false
		}
	}
	mapDark := true
	for _, h := range s.Map {
		if !s.Terminator.InShadow(h.Q) {
			mapDark = false
			break
		}
	}
	allStranded := true
	anyStranded := false
	for _, r := range s.Rovers {
		if r.State == RoverStranded {
			anyStranded = true
		} else {
			allStranded = false
		}
	}
	if len(s.Rovers) == 0 {
		allStranded = s.R().State == RoverStranded
		anyStranded = allStranded
	}
	_ = anyStranded
	if allStranded || allDone || mapDark || s.T >= GameDurationTargetSec {
		s.finish()
	}
}

func (s *GameState) finish() {
	if s.Status == StatusFinished {
		return
	}
	s.Status = StatusFinished
	for i := range s.Contracts {
		c := &s.Contracts[i]
		switch c.Status {
		case ContractQueued, ContractAccepted, ContractInTransit:
			if s.Terminator.InShadow(s.R().Q) {
				c.Status = ContractLostShadow
			} else {
				c.Status = ContractFailed
			}
		}
	}
	switch {
	case s.ColonyScore >= ColonyWinThreshold && s.EarthScore >= EarthPyrrhicBelow:
		s.Outcome = OutcomeSaved
	case s.ColonyScore >= ColonyWinThreshold:
		s.Outcome = OutcomePyrrhic
	default:
		s.Outcome = OutcomeLost
	}
	s.emit("game_over", map[string]any{"outcome": string(s.Outcome)})
}

func (s *GameState) emitOnce(kind string, payload map[string]any) {
	for _, e := range s.Events {
		if e.Kind == kind {
			return
		}
	}
	s.emit(kind, payload)
}
