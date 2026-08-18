package sim

func (s *GameState) Accept(id string) error {
	if s.Status == StatusFinished {
		return errGameOver
	}
	idx := -1
	for i, c := range s.Contracts {
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
	if reason := s.acceptBlocked(s.Contracts[idx]); reason != "" {
		s.reject(reason, id)
		return errConflict
	}
	s.Contracts[idx].Status = ContractAccepted
	s.Contracts[idx].AssignedTo = s.R().Type
	s.emit("contract_accepted", map[string]any{"contractId": id, "rover": string(s.R().Type)})
	return nil
}

func (s *GameState) acceptBlocked(c Contract) string {
	if s.activeSlots() >= MaxActiveContracts {
		return "slots_full"
	}
	if s.R().Type == RoverSwift && c.Weight == WeightClassHeavy {
		return "swift_no_heavy"
	}
	if s.R().State == RoverStranded {
		return "stranded"
	}
	return ""
}

func (s *GameState) Dispatch(id string) error {
	if s.Status == StatusFinished {
		return errGameOver
	}
	idx := -1
	for i, c := range s.Contracts {
		if c.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return errNotFound
	}
	c := s.Contracts[idx]
	if c.AssignedTo != "" && c.AssignedTo != s.R().Type {
		s.reject("wrong_rover", id)
		return nil
	}
	if s.R().State == RoverStranded {
		s.reject("stranded", id)
		return nil
	}
	if c.Status == ContractQueued {
		if err := s.Accept(id); err != nil {
			return nil
		}
		c = s.Contracts[idx]
	}
	if c.Status != ContractAccepted && c.Status != ContractInTransit {
		return errConflict
	}
	r := s.R()
	target := c.Pickup
	if c.Status == ContractInTransit || HexID(r.Q, r.R) == c.Pickup {
		s.rescueDropoffs()
		c = s.Contracts[idx]
		target = c.Dropoff
	}
	return s.GoTo(target)
}

func (s *GameState) Reroute(ids []string) error {
	r := s.R()
	if s.Status == StatusFinished || r.State == RoverStranded {
		return errGameOver
	}
	if err := s.SetRoute(ids); err != nil {
		return err
	}
	s.emit("reroute", nil)
	if r.State != RoverMoving {
		return nil
	}
	if s.FreeReroutes > 0 {
		s.FreeReroutes--
		return nil
	}
	r.Battery -= r.MaxBattery * 0.15
	if s.hasCargo(CargoReactor) {
		r.Battery -= 15
	}
	if r.Battery < 0 {
		r.Battery = 0
	}
	return nil
}

func (s *GameState) Autonomy() (applied bool, reason string) {
	r := s.R()
	if s.Status == StatusFinished {
		return false, "game_over"
	}
	if s.AutonomyLeft <= 0 {
		return false, "no_charges"
	}
	if len(r.Path) == 0 {
		return false, "no_path"
	}
	goal := r.Path[len(r.Path)-1]
	from := Axial{Q: r.Q, R: r.R}
	if r.State == RoverMoving && len(r.Path) > 0 && r.Progress > 0.02 {
		from = r.Path[0]
	}
	alt := s.FindPath(from, goal)
	if len(alt) == 0 {
		return false, "no_safer_path"
	}
	if len(alt) > len(r.Path)+2 {
		return false, "no_safer_path"
	}
	if pathRisk(s, alt) >= pathRisk(s, r.Path) {
		return false, "no_safer_path"
	}
	if r.State == RoverMoving && len(r.Path) > 0 && r.Progress > 0.02 {
		r.Path = append([]Axial{r.Path[0]}, alt...)
	} else {
		r.Path = alt
	}
	s.AutonomyLeft--
	s.emit("autonomy_used", map[string]any{"charges": s.AutonomyLeft})
	return true, ""
}

func pathRisk(s *GameState, path []Axial) int {
	risk := 0
	t := s.Terminator
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
		edge := cost / s.roverSpeed()
		t = t.Advance(edge)
	}
	return risk
}
