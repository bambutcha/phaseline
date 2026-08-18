package sim

import "math"

func (s *GameState) WeightMod() float64 {
	hasHeavy, hasMedium := false, false
	for _, c := range s.Contracts {
		if c.Status != ContractAccepted && c.Status != ContractInTransit {
			continue
		}
		switch c.Weight {
		case WeightClassHeavy:
			hasHeavy = true
		case WeightClassMedium:
			hasMedium = true
		}
	}
	if hasHeavy {
		return WeightHeavy
	}
	if hasMedium {
		return WeightMedium
	}
	return WeightLight
}

func terrainMod(rover RoverType, tile HexType) float64 {
	switch {
	case rover == RoverSwift && tile == TypeRidge:
		return 1.1
	case rover == RoverHauler && tile == TypeCrater:
		return 1.1
	default:
		return 1.0
	}
}

func shadowMod(inShadow bool) float64 {
	if inShadow {
		return ShadowMod
	}
	return 1.0
}

func (s *GameState) MoveCost(hex Hex) float64 {
	return s.moveCostAt(hex, s.WeightMod(), s.Terminator.InShadow(hex.Q))
}

func (s *GameState) moveCostAt(hex Hex, weightMod float64, inShadow bool) float64 {
	if hex.Impassable {
		return math.Inf(1)
	}
	return BaseMoveCost(hex.Type) * weightMod * terrainMod(s.Rover.Type, hex.Type) * shadowMod(inShadow)
}

func (s *GameState) roverSpeed() float64 {
	sp := s.Rover.Speed
	if s.Rover.PanicLeft > 0 {
		return sp * 0.7
	}
	return sp
}

func minBaseCost() float64 {
	return BaseMoveCost(TypeBase)
}
