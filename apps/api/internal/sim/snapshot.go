package sim

import "sort"

type Snapshot struct {
	Seed            string     `json:"seed"`
	Status          GameStatus `json:"status"`
	Outcome         Outcome    `json:"outcome,omitempty"`
	T               float64    `json:"t"`
	Rover           RoverView  `json:"rover"`
	Map             MapView    `json:"map"`
	Contracts       []Contract `json:"contracts"`
	Crisis          CrisisView `json:"crisis"`
	ColonyScore     int        `json:"colonyScore"`
	EarthScore      int        `json:"earthScore"`
	AutonomyCharges int        `json:"autonomyCharges"`
	Events          []Event    `json:"events,omitempty"`
}

type RoverView struct {
	Type     RoverType   `json:"type"`
	Hex      string      `json:"hex"`
	Q        int         `json:"q"`
	R        int         `json:"r"`
	Progress float64     `json:"progress"`
	Battery  float64     `json:"battery"`
	Max      float64     `json:"maxBattery"`
	State    RoverState  `json:"state"`
	Cargo    []CargoType `json:"cargo"`
	Path     []string    `json:"path"`
}

type MapView struct {
	Hexes      []HexView  `json:"hexes"`
	Terminator Terminator `json:"terminator"`
}

type HexView struct {
	ID         string  `json:"id"`
	Q          int     `json:"q"`
	R          int     `json:"r"`
	Type       HexType `json:"type"`
	Impassable bool    `json:"impassable"`
	InShadow   bool    `json:"inShadow"`
	PhaseETA   float64 `json:"phaseEta"`
}

type CrisisView struct {
	Kind    string  `json:"kind"`
	FiresAt float64 `json:"firesAt"`
	Fired   bool    `json:"fired"`
}

func (s *GameState) Snapshot() Snapshot {
	ids := make([]string, 0, len(s.Map))
	for id := range s.Map {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	hexes := make([]HexView, 0, len(ids))
	for _, id := range ids {
		h := s.Map[id]
		hexes = append(hexes, HexView{
			ID: id, Q: h.Q, R: h.R, Type: h.Type, Impassable: h.Impassable,
			InShadow: s.Terminator.InShadow(h.Q),
			PhaseETA: s.Terminator.PhaseETA(h.Q),
		})
	}
	path := make([]string, 0, len(s.Rover.Path))
	for _, a := range s.Rover.Path {
		path = append(path, a.ID())
	}
	return Snapshot{
		Seed:    s.Seed,
		Status:  s.Status,
		Outcome: s.Outcome,
		T:       round1(s.T),
		Rover: RoverView{
			Type: s.Rover.Type, Hex: HexID(s.Rover.Q, s.Rover.R),
			Q: s.Rover.Q, R: s.Rover.R, Progress: s.Rover.Progress,
			Battery: s.Rover.Battery, Max: s.Rover.MaxBattery,
			State: s.Rover.State, Cargo: s.Rover.Cargo, Path: path,
		},
		Map:         MapView{Hexes: hexes, Terminator: s.Terminator},
		Contracts:   append([]Contract(nil), s.Contracts...),
		Crisis:      CrisisView{Kind: s.CrisisKind, FiresAt: s.CrisisAt, Fired: s.CrisisFired},
		ColonyScore: s.ColonyScore, EarthScore: s.EarthScore,
		AutonomyCharges: s.AutonomyLeft,
		Events:          append([]Event(nil), s.Events...),
	}
}
