package sim

const (
	TickHz = 10
	TickDT = 0.1

	GameDurationTargetSec = 180.0
	ColonyWinThreshold    = 100
	EarthPyrrhicBelow     = 40
	AutonomyCharges       = 3
	RerouteFreeCount      = 1
	ContractsPerGame      = 6
	MaxActiveContracts    = 2
	MapHexCount           = 30
	MapCols               = 6
	MapRows               = 5

	TerminatorSpeed = 0.052

	SwiftSpeed      = 0.52
	HaulerSpeed     = 0.36
	SwiftBattery    = 80.0
	HaulerBattery   = 105.0
	SwiftIdleDrain  = 1.38
	HaulerIdleDrain = 1.05
	SolarGain       = 3.0
	BaseRecharge    = 4.6
	MoveDrainMult   = 1.72

	WeightLight  = 1.00
	WeightMedium = 1.15
	WeightHeavy  = 1.35
	ShadowMod    = 1.5
)

const (
	TypeBase         HexType = "base"
	TypeRegolith     HexType = "regolith"
	TypeSolarPlateau HexType = "solar_plateau"
	TypeDustField    HexType = "dust_field"
	TypeColdSink     HexType = "cold_sink"
	TypeCrater       HexType = "crater"
	TypeRidge        HexType = "ridge"
)

func BaseMoveCost(t HexType) float64 {
	switch t {
	case TypeBase:
		return 0.8
	case TypeRegolith:
		return 1.0
	case TypeSolarPlateau:
		return 0.9
	case TypeDustField:
		return 1.1
	case TypeColdSink:
		return 1.2
	case TypeCrater:
		return 1.4
	case TypeRidge:
		return 1.5
	default:
		return 1.0
	}
}
