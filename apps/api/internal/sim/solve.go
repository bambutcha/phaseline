package sim

import "math"

func deliverableCeiling(s *GameState) int {
	sum := 0
	for _, c := range s.Contracts {
		sum += c.ColonyValue
	}
	return sum
}

func densityColony(seed string) int {
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
			bestID := ""
			bestScore := -1.0
			here := Axial{Q: r.Q, R: r.R}
			for _, c := range s.Contracts {
				if c.Status != ContractQueued {
					continue
				}
				if reason := s.acceptBlocked(c); reason != "" {
					continue
				}
				pick, err := ParseHexID(c.Pickup)
				if err != nil {
					continue
				}
				drop, err := ParseHexID(c.Dropoff)
				if err != nil {
					continue
				}
				trip := float64(CubeDistance(here, pick)+CubeDistance(pick, drop)) / math.Max(0.08, r.Speed)
				eta := s.Terminator.PhaseETA(pick.Q)
				if eta > 0 && eta < trip*0.45 {
					continue
				}
				sc := float64(c.ColonyValue) / (trip + 1)
				if sc > bestScore {
					bestScore = sc
					bestID = c.ID
				}
			}
			if bestID != "" {
				_ = s.Dispatch(bestID)
			} else {
				s.seekSalvage()
			}
		}
		s.Active = sel
		s.Tick(TickDT)
	}
	return s.ColonyScore
}

func searchColony(seed string) int {
	base := NewGame(seed, RoverSwift)
	type job struct {
		id    string
		heavy bool
		val   int
	}
	jobs := make([]job, 0, len(base.Contracts))
	for _, c := range base.Contracts {
		jobs = append(jobs, job{id: c.ID, heavy: c.Weight == WeightClassHeavy, val: c.ColonyValue})
	}
	best := densityColony(seed)
	if best >= ColonyWinThreshold {
		return best
	}
	if len(jobs) > 4 {
		jobs = jobs[:4]
	}
	var rec func(i int, swiftQ, haulerQ []string)
	rec = func(i int, swiftQ, haulerQ []string) {
		if i == len(jobs) {
			need := 0
			for _, id := range swiftQ {
				for _, j := range jobs {
					if j.id == id {
						need += j.val
					}
				}
			}
			for _, id := range haulerQ {
				for _, j := range jobs {
					if j.id == id {
						need += j.val
					}
				}
			}
			if need < ColonyWinThreshold && best >= ColonyWinThreshold {
				return
			}
			sc := runQueues(seed, swiftQ, haulerQ)
			if sc > best {
				best = sc
			}
			return
		}
		j := jobs[i]
		rec(i+1, swiftQ, haulerQ)
		if !j.heavy {
			rec(i+1, append(append([]string{}, swiftQ...), j.id), haulerQ)
		}
		rec(i+1, swiftQ, append(append([]string{}, haulerQ...), j.id))
	}
	rec(0, nil, nil)
	return best
}

func runQueues(seed string, swiftJobs, haulerJobs []string) int {
	s := NewGame(seed, RoverSwift)
	s.Status = StatusActive
	si, hi := 0, 0
	for s.Status == StatusActive {
		si = pumpQueue(s, 0, swiftJobs, si)
		hi = pumpQueue(s, 1, haulerJobs, hi)
		s.Tick(TickDT * 2)
	}
	return s.ColonyScore
}

func pumpQueue(s *GameState, roverIdx int, jobs []string, cursor int) int {
	if roverIdx < 0 || roverIdx >= len(s.Rovers) {
		return cursor
	}
	s.Active = roverIdx
	if s.R().State != RoverIdle {
		return cursor
	}
	for cursor < len(jobs) {
		id := jobs[cursor]
		cursor++
		_ = s.Dispatch(id)
		if s.R().State == RoverMoving || len(s.R().Path) > 0 {
			break
		}
	}
	if s.R().State == RoverIdle {
		s.seekSalvage()
	}
	return cursor
}

func (s *GameState) seekSalvage() {
	r := s.R()
	if r.State != RoverIdle {
		return
	}
	here := Axial{Q: r.Q, R: r.R}
	best := ""
	bestD := 1 << 20
	for _, sv := range s.Salvage {
		if sv.Status != SalvageAvailable {
			continue
		}
		ax, err := ParseHexID(sv.Hex)
		if err != nil || s.Terminator.InShadow(ax.Q) {
			continue
		}
		d := CubeDistance(here, ax)
		if d < bestD {
			bestD = d
			best = sv.Hex
		}
	}
	if best != "" {
		_ = s.GoTo(best)
	}
}
