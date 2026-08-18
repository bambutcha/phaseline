package sim

func (s *GameState) Accept(id string) error {
	if s.Status == StatusFinished {
		return errGameOver
	}
	active := 0
	idx := -1
	for i, c := range s.Contracts {
		if c.Status == ContractAccepted || c.Status == ContractInTransit {
			active++
		}
		if c.ID == id {
			idx = i
		}
	}
	if idx < 0 {
		return errNotFound
	}
	if s.Contracts[idx].Status != ContractQueued {
		return errConflict
	}
	if active >= MaxActiveContracts {
		return errConflict
	}
	s.Contracts[idx].Status = ContractAccepted
	s.emit("contract_accepted", map[string]any{"contractId": id})
	return nil
}

func (s *GameState) Reroute(ids []string) error {
	if s.Status == StatusFinished || s.Rover.State == RoverStranded {
		return errGameOver
	}
	if err := s.SetRoute(ids); err != nil {
		return err
	}
	s.emit("reroute", nil)
	if s.Rover.State != RoverMoving {
		return nil
	}
	if s.FreeReroutes > 0 {
		s.FreeReroutes--
		return nil
	}
	s.Rover.Battery -= s.Rover.MaxBattery * 0.15
	if s.hasCargo(CargoReactor) {
		s.Rover.Battery -= 15
	}
	if s.Rover.Battery < 0 {
		s.Rover.Battery = 0
	}
	return nil
}

func (s *GameState) Autonomy() (applied bool, reason string) {
	if s.Status == StatusFinished {
		return false, "game_over"
	}
	if s.AutonomyLeft <= 0 {
		return false, "no_charges"
	}
	if len(s.Rover.Path) == 0 {
		return false, "no_path"
	}
	goal := s.Rover.Path[len(s.Rover.Path)-1]
	from := Axial{Q: s.Rover.Q, R: s.Rover.R}
	alt := s.FindPath(from, goal)
	if len(alt) == 0 {
		return false, "no_safer_path"
	}
	if len(alt) > len(s.Rover.Path)+2 {
		return false, "no_safer_path"
	}
	if pathRisk(s, alt) >= pathRisk(s, s.Rover.Path) {
		return false, "no_safer_path"
	}
	s.Rover.Path = alt
	s.AutonomyLeft--
	s.emit("autonomy_used", map[string]any{"charges": s.AutonomyLeft})
	return true, ""
}

func pathRisk(s *GameState, path []Axial) int {
	risk := 0
	t := s.Terminator
	elapsed := 0.0
	for _, a := range path {
		h := s.Map[a.ID()]
		if h.Type == TypeCrater {
			risk += 3
		}
		if t.InShadow(h.Q) {
			risk += 2
		}
		if h.Type == TypeRidge {
			risk += 1
		}
		cost := s.moveCostAt(h, s.WeightMod(), t.InShadow(h.Q))
		elapsed = cost / s.roverSpeed()
		t = t.Advance(elapsed)
	}
	return risk
}
