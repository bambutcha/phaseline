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
	ContractFailed     ContractStatus = "failed"
	ContractExpired    ContractStatus = "expired"
	ContractLostShadow ContractStatus = "lost_to_shadow"
)

type RoverState string

const (
	RoverIdle     RoverState = "idle"
	RoverMoving   RoverState = "moving"
	RoverStranded RoverState = "stranded"
)

type Hex struct {
	Q          int
	R          int
	Type       HexType
	Impassable bool
}

func (h Hex) ID() string {
	return HexID(h.Q, h.R)
}

type Terminator struct {
	Pos       float64
	Speed     float64
	Direction string // east, west, ...
}

type Rover struct {
	Type       RoverType
	Q, R       int
	Progress   float64
	Path       []Hex
	Battery    float64
	MaxBattery float64
	Speed      float64
	IdleDrain  float64
	State      RoverState
	Cargo      []CargoType
}

type Contract struct {
	ID          string
	Title       string
	Cargo       CargoType
	Weight      Weight
	Pickup      string
	Dropoff     string
	ColonyValue int
	EarthValue  int
	Deadline    float64
	Status      ContractStatus
}

type GameState struct {
	Seed         string
	T            float64
	Map          map[string]Hex
	Terminator   Terminator
	Rover        Rover
	Contracts    []Contract
	ColonyScore  int
	EarthScore   int
	CrisisKind   string
	CrisisAt     float64
	CrisisFired  bool
	AutonomyLeft int
	FreeReroutes int
}
