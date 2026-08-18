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
	if s.Rovers != nil {
		cp.Rovers = append([]Rover(nil), s.Rovers...)
		for i := range cp.Rovers {
			if s.Rovers[i].Path != nil {
				cp.Rovers[i].Path = append([]Axial(nil), s.Rovers[i].Path...)
			}
			if s.Rovers[i].Cargo != nil {
				cp.Rovers[i].Cargo = append([]CargoType(nil), s.Rovers[i].Cargo...)
			}
		}
	}
	if s.Events != nil {
		cp.Events = append([]Event(nil), s.Events...)
	}
	if s.Ghost != nil {
		g := *s.Ghost
		cp.Ghost = &g
	}
	if s.LastReject != nil {
		r := *s.LastReject
		cp.LastReject = &r
	}
	return &cp
}

func (s *GameState) Predict(path []Axial) Prediction {
	cp := s.Clone()
	cp.Events = nil
	r := cp.R()
	r.Path = append([]Axial(nil), path...)
	if cp.Status != StatusActive {
		cp.Status = StatusActive
	}
	if len(path) > 0 {
		r.State = RoverMoving
		r.Progress = 0
	}
	for i := range cp.Rovers {
		if i == cp.Active {
			continue
		}
		cp.Rovers[i].State = RoverIdle
		cp.Rovers[i].Path = nil
	}
	pred := Prediction{
		EndBattery: r.Battery,
		MinBattery: r.Battery,
		Feasible:   true,
	}
	startT := cp.T
	for i := 0; i < 10000; i++ {
		cp.Tick(TickDT)
		cur := cp.R()
		if cur.Battery < pred.MinBattery {
			pred.MinBattery = cur.Battery
		}
		if cur.State != RoverMoving || len(cur.Path) == 0 {
			break
		}
		if cur.State == RoverStranded {
			break
		}
	}
	pred.ETASec = cp.T - startT
	pred.EndBattery = cp.R().Battery
	pred.Feasible = pred.MinBattery > 0 && cp.R().State != RoverStranded
	seen := map[string]bool{}
	t := s.Terminator
	for _, step := range path {
		hex := s.Map[step.ID()]
		cost := s.moveCostAt(hex, s.WeightMod(), t.InShadow(hex.Q))
		edge := cost / s.roverSpeed()
		t = t.Advance(edge)
		if t.InShadow(step.Q) && !seen[step.ID()] {
			pred.InShadowAt = append(pred.InShadowAt, step.ID())
			seen[step.ID()] = true
		}
	}
	return pred
}
