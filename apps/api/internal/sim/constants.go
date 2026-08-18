package sim

const (
	TickHz = 10
	TickDT = 0.1

	GameDurationTargetSec = 90.0
	ColonyWinThreshold    = 80
	EarthPyrrhicBelow     = 30
	AutonomyCharges       = 3
	RerouteFreeCount      = 1
	ContractsPerGame      = 6
	MaxActiveContracts    = 2
	MapHexCount           = 20
	MapCols               = 5
	MapRows               = 4

	TerminatorSpeed = 0.048

	SwiftSpeed      = 1.15
	HaulerSpeed     = 0.85
	SwiftBattery    = 100.0
	HaulerBattery   = 140.0
	SwiftIdleDrain  = 0.7
	HaulerIdleDrain = 0.5
	SolarGain       = 10.0
	BaseRecharge    = 12.0

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
