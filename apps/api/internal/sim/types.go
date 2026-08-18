package sim

type HexType string

type CargoType string

const (
	CargoO2        CargoType = "o2_filters"
	CargoCryo      CargoType = "cryo_samples"
	CargoCrew      CargoType = "crew_pod"
	CargoReactor   CargoType = "reactor_rod"
	CargoHelium3   CargoType = "helium3"
	CargoMedSeeds  CargoType = "med_seeds"
	CargoCommRelay CargoType = "comm_relay"
)

type Weight string

const (
	WeightClassLight  Weight = "light"
	WeightClassMedium Weight = "medium"
	WeightClassHeavy  Weight = "heavy"
)

type RoverType string

const (
	RoverSwift  RoverType = "swift"
	RoverHauler RoverType = "hauler"
)

type ContractStatus string

const (
	ContractQueued     ContractStatus = "queued"
	ContractAccepted   ContractStatus = "accepted"
	ContractInTransit  ContractStatus = "in_transit"
	ContractDelivered  ContractStatus = "delivered"
	ContractExpired    ContractStatus = "expired"
	ContractFailed     ContractStatus = "failed"
	ContractLostShadow ContractStatus = "lost_to_shadow"
)

type RoverState string

const (
	RoverIdle     RoverState = "idle"
	RoverMoving   RoverState = "moving"
	RoverStranded RoverState = "stranded"
)

type GameStatus string

const (
	StatusLobby    GameStatus = "lobby"
	StatusActive   GameStatus = "active"
	StatusFinished GameStatus = "finished"
)

type Outcome string

const (
	OutcomeNone    Outcome = ""
	OutcomeSaved   Outcome = "colony_saved"
	OutcomePyrrhic Outcome = "pyrrhic"
	OutcomeLost    Outcome = "signal_lost"
)

const (
	DirEast = "east"
	DirWest = "west"
)

type Hex struct {
	Q          int     `json:"q"`
	R          int     `json:"r"`
	Type       HexType `json:"type"`
	Impassable bool    `json:"impassable"`
}

func (h Hex) ID() string {
	return HexID(h.Q, h.R)
}

func (h Hex) Axial() Axial {
	return Axial{Q: h.Q, R: h.R}
}

type Terminator struct {
	Pos       float64 `json:"pos"`
	Speed     float64 `json:"speed"`
	Direction string  `json:"direction"`
}

type Rover struct {
	Type       RoverType   `json:"type"`
	Q, R       int         `json:"q"`
	Progress   float64     `json:"progress"`
	Path       []Axial     `json:"-"`
	Battery    float64     `json:"battery"`
	MaxBattery float64     `json:"maxBattery"`
	Speed      float64     `json:"speed"`
	IdleDrain  float64     `json:"idleDrain"`
	State      RoverState  `json:"state"`
	Cargo      []CargoType `json:"cargo"`
	SunIdle    float64     `json:"-"`
	PanicLeft  int         `json:"-"`
	Reversing  bool        `json:"-"`
	ReverseTo  Axial       `json:"-"`
}

type Contract struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Cargo       CargoType      `json:"cargoType"`
	Weight      Weight         `json:"weight"`
	Pickup      string         `json:"pickup"`
	Dropoff     string         `json:"dropoff"`
	ColonyValue int            `json:"colonyValue"`
	EarthValue  int            `json:"earthValue"`
	Reward      int            `json:"reward"`
	Risk        string         `json:"risk"`
	Urgency     string         `json:"urgency"`
	Deadline    float64        `json:"deadline"`
	Status      ContractStatus `json:"status"`
	AssignedTo  RoverType      `json:"assignedTo,omitempty"`
	Impossible  bool           `json:"impossible,omitempty"`
}

type Event struct {
	T       float64        `json:"t"`
	Kind    string         `json:"kind"`
	Payload map[string]any `json:"payload,omitempty"`
}

type GameState struct {
	Seed         string
	Layout       string
	Status       GameStatus
	Outcome      Outcome
	T            float64
	Map          map[string]Hex
	Terminator   Terminator
	Rovers       []Rover
	Active       int
	Contracts    []Contract
	ColonyScore  int
	EarthScore   int
	CrisisKind   string
	CrisisAt     float64
	CrisisFired  bool
	AutonomyLeft int
	FreeReroutes int
	DustStorm    bool
	FlareActive  bool
	CommUntil    float64
	Events       []Event
	Ghost        *GhostReplay
	LastReject   *RejectView
	EndReason    string
	noChain      bool
}

type RejectView struct {
	Reason     string `json:"reason"`
	ContractID string `json:"contractId,omitempty"`
}

func (s *GameState) R() *Rover {
	if len(s.Rovers) == 0 {
		s.Rovers = []Rover{NewRover(RoverSwift, 0, 0)}
		s.Active = 0
	}
	if s.Active < 0 || s.Active >= len(s.Rovers) {
		s.Active = 0
	}
	return &s.Rovers[s.Active]
}

func (s *GameState) Hex(q, r int) (Hex, bool) {
	h, ok := s.Map[HexID(q, r)]
	return h, ok
}

func (s *GameState) RoverHex() Hex {
	r := s.R()
	h, ok := s.Hex(r.Q, r.R)
	if !ok {
		return Hex{Q: r.Q, R: r.R, Type: TypeRegolith}
	}
	return h
}

func (s *GameState) SelectRover(t RoverType) {
	for i := range s.Rovers {
		if s.Rovers[i].Type == t {
			s.Active = i
			return
		}
	}
}
