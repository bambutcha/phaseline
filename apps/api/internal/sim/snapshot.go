package sim

import "sort"

type Snapshot struct {
	Seed            string     `json:"seed"`
	Status          GameStatus `json:"status"`
	Outcome         Outcome    `json:"outcome,omitempty"`
	T               float64    `json:"t"`
	Rover           RoverView    `json:"rover"`
	Rovers          []RoverView  `json:"rovers"`
	ActiveRover     RoverType    `json:"activeRover"`
	Map             MapView    `json:"map"`
	Contracts       []Contract `json:"contracts"`
	Crisis          CrisisView `json:"crisis"`
	ColonyScore     int        `json:"colonyScore"`
	EarthScore      int        `json:"earthScore"`
	AutonomyCharges int         `json:"autonomyCharges"`
	Events          []Event     `json:"events,omitempty"`
	RoutePreview    *Prediction `json:"routePreview,omitempty"`
	Ghost           *GhostReplay `json:"ghost,omitempty"`
	Goal            GoalView     `json:"goal"`
	Reject          *RejectView  `json:"reject,omitempty"`
}

type GoalView struct {
	ColonyNeed int     `json:"colonyNeed"`
	EarthSafe  int     `json:"earthSafe"`
	Duration   float64 `json:"duration"`
}

type RoverView struct {
	Type      RoverType   `json:"type"`
	Name      string      `json:"name"`
	Hex       string      `json:"hex"`
	Q         int         `json:"q"`
	R         int         `json:"r"`
	Progress  float64     `json:"progress"`
	Battery   float64     `json:"battery"`
	Max       float64     `json:"maxBattery"`
	State     RoverState  `json:"state"`
	Cargo     []CargoType `json:"cargo"`
	Path      []string    `json:"path"`
	SlotsUsed int         `json:"slotsUsed"`
	SlotsMax  int         `json:"slotsMax"`
	CanHeavy  bool        `json:"canHeavy"`
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
	Risk       string  `json:"risk"`
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
			Risk:     hexRisk(h),
		})
	}
	views := make([]RoverView, 0, len(s.Rovers))
	for i := range s.Rovers {
		views = append(views, s.roverView(&s.Rovers[i]))
	}
	snap := Snapshot{
		Seed:    s.Seed,
		Status:  s.Status,
		Outcome: s.Outcome,
		T:       round1(s.T),
		Rover:   s.roverView(s.R()),
		Rovers:  views,
		ActiveRover: s.R().Type,
		Map:             MapView{Hexes: hexes, Terminator: s.Terminator},
		Contracts:       append([]Contract(nil), s.Contracts...),
		Crisis:          CrisisView{Kind: s.CrisisKind, FiresAt: s.CrisisAt, Fired: s.CrisisFired},
		ColonyScore:     s.ColonyScore,
		EarthScore:      s.EarthScore,
		AutonomyCharges: s.AutonomyLeft,
		Events:          append([]Event(nil), s.Events...),
		Ghost:           s.Ghost,
		Goal:            GoalView{ColonyNeed: ColonyWinThreshold, EarthSafe: EarthPyrrhicBelow, Duration: GameDurationTargetSec},
		Reject:          s.LastReject,
	}
	if len(s.R().Path) > 0 && s.Status != StatusFinished {
		pred := s.Predict(append([]Axial(nil), s.R().Path...))
		snap.RoutePreview = &pred
	}
	return snap
}

func roverPath(r *Rover) []string {
	path := make([]string, 0, len(r.Path))
	for _, a := range r.Path {
		path = append(path, a.ID())
	}
	return path
}

func (s *GameState) roverView(r *Rover) RoverView {
	return RoverView{
		Type: r.Type, Name: roverName(r.Type),
		Hex: HexID(r.Q, r.R), Q: r.Q, R: r.R, Progress: r.Progress,
		Battery: r.Battery, Max: r.MaxBattery, State: r.State, Cargo: r.Cargo,
		Path: roverPath(r), SlotsUsed: s.slotsFor(r.Type), SlotsMax: MaxActiveContracts,
		CanHeavy: r.Type == RoverHauler,
	}
}

func roverName(t RoverType) string {
	if t == RoverHauler {
		return "Hauler"
	}
	return "Swift"
}

func hexRisk(h Hex) string {
	switch h.Type {
	case TypeCrater, TypeColdSink:
		return "high"
	case TypeRidge, TypeDustField:
		return "medium"
	default:
		return "low"
	}
}

func (s *GameState) activeSlots() int {
	return s.slotsFor(s.R().Type)
}

func (s *GameState) slotsFor(t RoverType) int {
	n := 0
	for _, c := range s.Contracts {
		if c.AssignedTo != t {
			continue
		}
		if c.Status == ContractAccepted || c.Status == ContractInTransit {
			n++
		}
	}
	return n
}
