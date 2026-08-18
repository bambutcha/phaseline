package sim

const (
	TickHz = 10
	TickDT = 0.1

	GameDurationTargetSec = 240.0
	ColonyWinThreshold    = 100
	EarthPyrrhicBelow     = 40
	AutonomyCharges       = 3
	RerouteFreeCount      = 1
	ContractsPerGame      = 6
	MaxActiveContracts    = 2
	MapHexCount           = 12

	TerminatorSpeed = 0.015 // hex projection units per second

	SwiftSpeed      = 0.08
	HaulerSpeed     = 0.05
	SwiftBattery    = 100.0
	HaulerBattery   = 140.0
	SwiftIdleDrain  = 2.0
	HaulerIdleDrain = 1.5
	SolarGain       = 8.0

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
