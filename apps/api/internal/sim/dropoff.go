package sim

import "sort"

func (s *GameState) farBaseID() string {
	if s.Terminator.Direction == DirWest {
		return HexID(0, MapRows/2)
	}
	return HexID(MapCols-1, MapRows/2)
}

func (s *GameState) litDropoff() string {
	prefer := []string{s.farBaseID()}
	ids := make([]string, 0)
	for id, h := range s.Map {
		if h.Type == TypeBase {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	prefer = append(prefer, ids...)
	seen := map[string]bool{}
	for _, id := range prefer {
		if seen[id] {
			continue
		}
		seen[id] = true
		h, ok := s.Map[id]
		if !ok || h.Impassable || s.Terminator.InShadow(h.Q) {
			continue
		}
		return id
	}
	return s.farBaseID()
}

func (s *GameState) rescueDropoffs() {
	dest := s.litDropoff()
	if dest == "" {
		return
	}
	for i := range s.Contracts {
		c := &s.Contracts[i]
		switch c.Status {
		case ContractQueued, ContractAccepted, ContractInTransit:
		default:
			continue
		}
		ax, err := ParseHexID(c.Dropoff)
		if err != nil || c.Dropoff == dest {
			continue
		}
		if !s.Terminator.InShadow(ax.Q) {
			continue
		}
		old := c.Dropoff
		c.Dropoff = dest
		s.emit("dropoff_moved", map[string]any{"contractId": c.ID, "from": old, "to": dest})
	}
}

func (s *GameState) retargetCargoRoutes() {
	sel := s.Active
	for i := range s.Rovers {
		s.Active = i
		r := s.R()
		if r.State == RoverStranded {
			continue
		}
		here := HexID(r.Q, r.R)
		for _, c := range s.Contracts {
			if c.AssignedTo != r.Type || c.Status != ContractInTransit {
				continue
			}
			if c.Dropoff == "" || c.Dropoff == here {
				continue
			}
			if len(r.Path) > 0 && r.Path[len(r.Path)-1].ID() == c.Dropoff {
				continue
			}
			_ = s.GoTo(c.Dropoff)
			break
		}
	}
	s.Active = sel
}
