package sim

type Prediction struct {
	ETASec     float64  `json:"etaSec"`
	EndBattery float64  `json:"predictedBattery"`
	MinBattery float64  `json:"minBattery"`
	Feasible   bool     `json:"feasible"`
	InShadowAt []string `json:"inShadowAt"`
}

func (s *GameState) Clone() *GameState {
	cp := *s
	cp.Map = make(map[string]Hex, len(s.Map))
	for k, v := range s.Map {
		cp.Map[k] = v
	}
	if s.Contracts != nil {
		cp.Contracts = append([]Contract(nil), s.Contracts...)
	}
	if s.Rover.Path != nil {
		cp.Rover.Path = append([]Axial(nil), s.Rover.Path...)
	}
	if s.Rover.Cargo != nil {
		cp.Rover.Cargo = append([]CargoType(nil), s.Rover.Cargo...)
	}
	if s.Events != nil {
		cp.Events = append([]Event(nil), s.Events...)
	}
	return &cp
}

func (s *GameState) Predict(path []Axial) Prediction {
	cp := s.Clone()
	cp.Events = nil
	cp.Rover.Path = append([]Axial(nil), path...)
	if cp.Status != StatusActive {
		cp.Status = StatusActive
	}
	if len(path) > 0 {
		cp.Rover.State = RoverMoving
		cp.Rover.Progress = 0
	}
	pred := Prediction{
		EndBattery: cp.Rover.Battery,
		MinBattery: cp.Rover.Battery,
		Feasible:   true,
	}
	startT := cp.T
	for i := 0; i < 10000; i++ {
		cp.Tick(TickDT)
		if cp.Rover.Battery < pred.MinBattery {
			pred.MinBattery = cp.Rover.Battery
		}
		if cp.Rover.State != RoverMoving || len(cp.Rover.Path) == 0 {
			break
		}
		if cp.Rover.State == RoverStranded {
			break
		}
	}
	pred.ETASec = cp.T - startT
	pred.EndBattery = cp.Rover.Battery
	pred.Feasible = pred.MinBattery > 0 && cp.Rover.State != RoverStranded
	seen := map[string]bool{}
	at := Axial{Q: s.Rover.Q, R: s.Rover.R}
	t := s.Terminator
	elapsed := 0.0
	for _, step := range path {
		hex := s.Map[step.ID()]
		cost := s.moveCostAt(hex, s.WeightMod(), t.InShadow(hex.Q))
		edge := cost / s.roverSpeed()
		elapsed += edge
		t = t.Advance(edge)
		if t.InShadow(step.Q) && !seen[step.ID()] {
			pred.InShadowAt = append(pred.InShadowAt, step.ID())
			seen[step.ID()] = true
		}
		at = step
	}
	_ = at
	return pred
}
