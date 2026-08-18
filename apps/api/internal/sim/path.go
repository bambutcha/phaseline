package sim

import "math"

// FindPath returns hexes after start, inclusive of goal. Empty if start==goal or unreachable.
func (s *GameState) FindPath(from, to Axial) []Axial {
	if from == to {
		return nil
	}
	if _, ok := s.Map[to.ID()]; !ok {
		return nil
	}
	if s.Map[to.ID()].Impassable {
		return nil
	}

	type item struct {
		at Axial
		g  float64
	}
	gScore := map[string]float64{from.ID(): 0}
	came := map[string]Axial{}
	open := []item{{at: from, g: 0}}

	heuristic := func(a Axial) float64 {
		return float64(CubeDistance(a, to)) * minBaseCost() / s.R().Speed
	}

	for len(open) > 0 {
		best := 0
		bestF := open[0].g + heuristic(open[0].at)
		for i := 1; i < len(open); i++ {
			f := open[i].g + heuristic(open[i].at)
			if f < bestF {
				bestF = f
				best = i
			}
		}
		cur := open[best]
		open = append(open[:best], open[best+1:]...)
		if cur.at == to {
			return reconstruct(came, from, to)
		}
		term := s.Terminator.Advance(cur.g)
		for _, d := range Neighbors {
			nb := Axial{Q: cur.at.Q + d.Q, R: cur.at.R + d.R}
			hex, ok := s.Map[nb.ID()]
			if !ok || hex.Impassable {
				continue
			}
			cost := s.moveCostAt(hex, s.WeightMod(), term.InShadow(hex.Q))
			if math.IsInf(cost, 1) {
				continue
			}
			edgeTime := cost / s.roverSpeed()
			if edgeTime <= 0 {
				edgeTime = TickDT
			}
			ng := cur.g + edgeTime
			id := nb.ID()
			if prev, seen := gScore[id]; seen && ng >= prev {
				continue
			}
			gScore[id] = ng
			came[id] = cur.at
			open = append(open, item{at: nb, g: ng})
		}
	}
	return nil
}

func reconstruct(came map[string]Axial, from, to Axial) []Axial {
	var rev []Axial
	cur := to
	for cur != from {
		rev = append(rev, cur)
		prev, ok := came[cur.ID()]
		if !ok {
			return nil
		}
		cur = prev
	}
	out := make([]Axial, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		out = append(out, rev[i])
	}
	return out
}

func (s *GameState) SetRoute(ids []string) error {
	r := s.R()
	if r.State == RoverStranded || s.Status == StatusFinished {
		return errGameOver
	}
	if len(ids) == 0 {
		r.Path = nil
		return nil
	}
	prev := Axial{Q: r.Q, R: r.R}
	path := make([]Axial, 0, len(ids))
	for _, id := range ids {
		ax, err := ParseHexID(id)
		if err != nil {
			return err
		}
		hex, ok := s.Map[ax.ID()]
		if !ok || hex.Impassable {
			return errInvalidRoute
		}
		if CubeDistance(prev, ax) != 1 {
			return errInvalidRoute
		}
		path = append(path, ax)
		prev = ax
	}
	r.Path = path
	return nil
}

func (s *GameState) GoTo(hexID string) error {
	r := s.R()
	if r.State == RoverStranded || s.Status == StatusFinished {
		return errGameOver
	}
	to, err := ParseHexID(hexID)
	if err != nil {
		return err
	}

	from := Axial{Q: r.Q, R: r.R}
	locked := r.State == RoverMoving && r.Progress > 0.02 && len(r.Path) > 0 && !r.Reversing
	var through Axial
	if locked {
		through = r.Path[0]
	}

	if from == to {
		if locked || (r.State == RoverMoving && r.Progress > 0.02 && (len(r.Path) > 0 || r.Reversing)) {
			if !r.Reversing && len(r.Path) > 0 {
				r.ReverseTo = r.Path[0]
			}
			r.Reversing = true
			r.Path = nil
			return nil
		}
		r.Path = nil
		r.Reversing = false
		r.State = RoverIdle
		r.Progress = 0
		return nil
	}

	var path []Axial
	if locked {
		if through == to {
			path = []Axial{through}
		} else {
			rest := s.FindPath(through, to)
			if len(rest) == 0 {
				s.reject("no_path", "")
				return errInvalidRoute
			}
			path = append([]Axial{through}, rest...)
		}
	} else {
		path = s.FindPath(from, to)
		if len(path) == 0 {
			s.reject("no_path", "")
			return errInvalidRoute
		}
	}

	pred := s.Predict(path)
	if !pred.Feasible {
		s.reject("battery", "")
		if !locked {
			r.Path = path
		}
		return nil
	}
	s.LastReject = nil
	r.Reversing = false
	r.Path = path
	return s.startMoving(locked)
}

func (s *GameState) Deploy() error {
	return s.startMoving(false)
}

func (s *GameState) startMoving(keepProgress bool) error {
	r := s.R()
	if s.Status == StatusFinished || r.State == RoverStranded {
		return errGameOver
	}
	if len(r.Path) == 0 {
		return errInvalidRoute
	}
	r.State = RoverMoving
	if !keepProgress {
		r.Progress = 0
	}
	if s.Status == StatusLobby {
		s.Status = StatusActive
		s.emit("game_start", nil)
	}
	s.emit("deploy", map[string]any{"rover": string(r.Type), "hexId": HexID(r.Q, r.R)})
	return nil
}
