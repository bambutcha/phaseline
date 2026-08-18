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

	if s.Rover.State == RoverMoving {
		s.tickMove(dt)
	}

	s.tickBattery(dt)
	s.tickCargo(dt)
	s.tickCrisis()
	s.checkEnd()

	if s.Rover.Battery <= 0 {
		s.Rover.Battery = 0
		if s.Rover.State == RoverMoving || s.Rover.State == RoverIdle {
			if s.Rover.State == RoverMoving {
				s.failCarried(ContractFailed)
				s.Rover.State = RoverStranded
				s.Rover.Path = nil
				s.emit("stranded", map[string]any{"hexId": HexID(s.Rover.Q, s.Rover.R)})
			}
		}
	}

	if s.Rover.Battery <= s.Rover.MaxBattery*0.2 && s.Rover.Battery > 0 {
		s.emitOnce("battery_low", map[string]any{"battery": s.Rover.Battery})
	}

	return s.Events[start:]
}

func (s *GameState) tickMove(dt float64) {
	if len(s.Rover.Path) == 0 {
		s.Rover.State = RoverIdle
		s.Rover.Progress = 0
		return
	}
	next := s.Rover.Path[0]
	hex, ok := s.Map[next.ID()]
	if !ok || hex.Impassable {
		s.Rover.State = RoverIdle
		s.Rover.Path = nil
		s.Rover.Progress = 0
		return
	}
	cost := s.MoveCost(hex)
	if math.IsInf(cost, 1) {
		s.Rover.State = RoverIdle
		return
	}
	speed := s.roverSpeed()
	edgeTime := cost / speed
	if edgeTime <= 0 {
		edgeTime = TickDT
	}
	s.Rover.Progress += dt / edgeTime
	if s.Rover.Progress < 1 {
		return
	}

	prevShadow := s.InShadow(s.Rover.Q, s.Rover.R)
	s.Rover.Battery -= cost
	s.Rover.Q, s.Rover.R = next.Q, next.R
	s.Rover.Path = s.Rover.Path[1:]
	s.Rover.Progress = 0
	s.Rover.SunIdle = 0
	if s.Rover.PanicLeft > 0 {
		s.Rover.PanicLeft--
	}
	s.emit("hex_entered", map[string]any{"hexId": next.ID()})

	nowShadow := s.InShadow(s.Rover.Q, s.Rover.R)
	if nowShadow && !prevShadow {
		s.emit("entered_shadow", map[string]any{"hexId": next.ID()})
	}
	if !nowShadow && prevShadow {
		s.emit("left_shadow", map[string]any{"hexId": next.ID()})
	}

	if hex.Type == TypeCrater || hex.Type == TypeRidge {
		if s.hasCargo(CargoCrew) {
			s.Rover.PanicLeft = 1
		}
	}

	s.tryPickupDrop()

	if len(s.Rover.Path) == 0 {
		s.Rover.State = RoverIdle
	}
	if s.Rover.Battery <= 0 {
		s.Rover.Battery = 0
	}
}

func (s *GameState) tickBattery(dt float64) {
	here := s.RoverHex()
	inShadow := s.InShadow(s.Rover.Q, s.Rover.R)
	drain := s.Rover.IdleDrain * dt
	if inShadow {
		drain *= 2
		if s.hasCargo(CargoO2) {
			drain *= 2
		}
	}
	s.Rover.Battery -= drain

	if !inShadow && here.Type == TypeSolarPlateau && !s.FlareActive {
		s.Rover.Battery += SolarGain * dt
	}
	if s.Rover.Battery > s.Rover.MaxBattery {
		s.Rover.Battery = s.Rover.MaxBattery
	}
	if s.Rover.State == RoverIdle && !inShadow {
		s.Rover.SunIdle += dt
	} else if inShadow {
		s.Rover.SunIdle = 0
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
					s.dropCargo(c.Cargo)
					s.emit("contract_failed", map[string]any{"contractId": c.ID, "reason": "expired"})
				}
			}
		}
		if c.Status == ContractInTransit && c.Cargo == CargoCryo && s.Rover.SunIdle >= 1 {
			c.Status = ContractFailed
			s.dropCargo(CargoCryo)
			s.emit("contract_failed", map[string]any{"contractId": c.ID, "reason": "spoil"})
		}
	}
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
	roverID := HexID(s.Rover.Q, s.Rover.R)
	for _, id := range ids {
		h := s.Map[id]
		if h.Type == TypeCrater && !h.Impassable && id != roverID {
			h.Impassable = true
			s.Map[id] = h
			return
		}
	}
}

func (s *GameState) addVIP() {
	drop := HexID(s.Rover.Q, s.Rover.R)
	ids := make([]string, 0, len(s.Map))
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
		Pickup:      HexID(s.Rover.Q, s.Rover.R),
		Dropoff:     drop,
		ColonyValue: 40,
		EarthValue:  10,
		Deadline:    45,
		Status:      ContractQueued,
	})
}

func (s *GameState) tryPickupDrop() {
	here := HexID(s.Rover.Q, s.Rover.R)
	for i := range s.Contracts {
		c := &s.Contracts[i]
		if c.Status == ContractAccepted && c.Pickup == here {
			c.Status = ContractInTransit
			s.Rover.Cargo = append(s.Rover.Cargo, c.Cargo)
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
			s.emit("deliver", map[string]any{"contractId": c.ID, "hexId": here})
		}
	}
}

func (s *GameState) hasCargo(t CargoType) bool {
	for _, c := range s.Rover.Cargo {
		if c == t {
			return true
		}
	}
	return false
}

func (s *GameState) dropCargo(t CargoType) {
	out := s.Rover.Cargo[:0]
	for _, c := range s.Rover.Cargo {
		if c != t {
			out = append(out, c)
		}
	}
	s.Rover.Cargo = out
}

func (s *GameState) failCarried(st ContractStatus) {
	for i := range s.Contracts {
		c := &s.Contracts[i]
		if c.Status == ContractInTransit || c.Status == ContractAccepted {
			if st == ContractFailed && s.InShadow(s.Rover.Q, s.Rover.R) {
				c.Status = ContractLostShadow
			} else {
				c.Status = st
			}
			s.emit("contract_failed", map[string]any{"contractId": c.ID})
		}
	}
	s.Rover.Cargo = nil
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
	if s.Rover.State == RoverStranded || allDone || mapDark || s.T >= GameDurationTargetSec {
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
			if s.Terminator.InShadow(s.Rover.Q) {
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
