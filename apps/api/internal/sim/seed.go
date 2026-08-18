package sim

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const seedAlphabet = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

type cargoSpec struct {
	cargo    CargoType
	title    string
	weight   Weight
	colony   int
	earth    int
	deadline float64
}

var cargoPool = []cargoSpec{
	{CargoO2, "O₂-фильтры — Hab Block C", WeightClassMedium, 22, 8, 0},
	{CargoO2, "Скрубберы CO₂ — шлюз 2", WeightClassMedium, 20, 6, 0},
	{CargoO2, "Вода — льды Шеклтона", WeightClassMedium, 18, 9, 0},
	{CargoCryo, "Криопробы — кратер Шиккард", WeightClassLight, 10, 28, 0},
	{CargoCrew, "Капсула экипажа — смена «Рассвет»", WeightClassMedium, 28, 4, 0},
	{CargoCrew, "Медик — отсек B", WeightClassMedium, 24, 5, 0},
	{CargoReactor, "Стержень реактора — ядро Alpha", WeightClassHeavy, 16, 30, 0},
	{CargoReactor, "Топливные стержни — южный реактор", WeightClassHeavy, 14, 24, 0},
	{CargoHelium3, "Helium-3 — контракт Земли", WeightClassHeavy, 14, 32, 0},
	{CargoMedSeeds, "Медсемена — оранжерея", WeightClassLight, 16, 10, 0},
	{CargoCommRelay, "Реле связи — пик хребта", WeightClassLight, 14, 12, 0},
	{CargoCommRelay, "Маяк — южный гребень", WeightClassLight, 12, 14, 0},
}

var layoutNames = []string{"mixed", "solar_belt", "ridge_wall", "crater_field", "dust_sea", "cold_front"}

var crisisKinds = []string{
	"dust_storm", "solar_flare", "cave_in", "vip_override", "comm_blackout",
}

func NormalizeSeed(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return randomSeed(time.Now().UnixNano())
	}
	if !strings.HasPrefix(s, "MCC-") {
		s = "MCC-" + s
	}
	return s
}

func randomSeed(n int64) string {
	r := rand.New(rand.NewSource(n))
	b := make([]byte, 4)
	for i := range b {
		b[i] = seedAlphabet[r.Intn(len(seedAlphabet))]
	}
	return "MCC-" + string(b)
}

func seedRNG(seed string) *rand.Rand {
	sum := sha256.Sum256([]byte(seed))
	v := int64(binary.LittleEndian.Uint64(sum[:8]))
	return rand.New(rand.NewSource(v))
}

func NewRover(t RoverType, q, r int) Rover {
	if t != RoverHauler {
		t = RoverSwift
	}
	rvr := Rover{
		Type: t, Q: q, R: r, State: RoverIdle,
	}
	if t == RoverHauler {
		rvr.Battery = HaulerBattery
		rvr.MaxBattery = HaulerBattery
		rvr.Speed = HaulerSpeed
		rvr.IdleDrain = HaulerIdleDrain
	} else {
		rvr.Battery = SwiftBattery
		rvr.MaxBattery = SwiftBattery
		rvr.Speed = SwiftSpeed
		rvr.IdleDrain = SwiftIdleDrain
	}
	return rvr
}

func NewGame(seed string, rover RoverType) *GameState {
	seed = NormalizeSeed(seed)
	rng := seedRNG(seed)
	dir := DirEast
	if rng.Intn(2) == 1 {
		dir = DirWest
	}

	layout := layoutNames[rng.Intn(len(layoutNames))]
	hexes, ids := paintMap(rng, layout)

	westBase := Hex{Q: 0, R: MapRows / 2, Type: TypeBase}
	eastBase := Hex{Q: MapCols - 1, R: MapRows / 2, Type: TypeBase}
	hexes[westBase.ID()] = westBase
	hexes[eastBase.ID()] = eastBase

	start := westBase.Axial()
	if dir == DirWest {
		start = eastBase.Axial()
	}
	hq, hr := start.Q, start.R
	if start.R+1 < MapRows {
		hr = start.R + 1
	} else if start.R > 0 {
		hr = start.R - 1
	}

	span := float64(MapCols)
	term := Terminator{
		Speed:     span / (GameDurationTargetSec * 0.64) * (0.92 + rng.Float64()*0.16),
		Direction: dir,
	}
	if dir == DirWest {
		term.Pos = float64(MapCols) - 0.5
	} else {
		term.Pos = -0.5
	}

	crisis := crisisKinds[rng.Intn(len(crisisKinds))]
	jitter := rng.Float64()*20 - 10
	crisisAt := 0.28*GameDurationTargetSec + jitter

	swift := NewRover(RoverSwift, start.Q, start.R)
	hauler := NewRover(RoverHauler, hq, hr)
	active := 0
	if rover == RoverHauler {
		active = 1
	}

	g := &GameState{
		Seed:         seed,
		Layout:       layout,
		Status:       StatusLobby,
		Map:          hexes,
		Terminator:   term,
		Rovers:       []Rover{swift, hauler},
		Active:       active,
		CrisisKind:   crisis,
		CrisisAt:     crisisAt,
		AutonomyLeft: AutonomyCharges,
		FreeReroutes: RerouteFreeCount,
	}
	g.Contracts = g.rollContracts(rng, westBase.ID(), eastBase.ID(), ids)
	g.boostDeliverableCeiling()
	return g
}

func paintMap(rng *rand.Rand, layout string) (map[string]Hex, []string) {
	hexes := make(map[string]Hex, MapHexCount)
	ids := make([]string, 0, MapHexCount)
	for r := 0; r < MapRows; r++ {
		for q := 0; q < MapCols; q++ {
			h := Hex{Q: q, R: r, Type: terrainFor(rng, layout, q, r)}
			hexes[h.ID()] = h
			ids = append(ids, h.ID())
		}
	}
	return hexes, ids
}

func terrainFor(rng *rand.Rand, layout string, q, r int) HexType {
	switch layout {
	case "solar_belt":
		if r == MapRows/2 || r == MapRows/2-1 {
			return TypeSolarPlateau
		}
	case "ridge_wall":
		if q == MapCols/2 {
			return TypeRidge
		}
	case "crater_field":
		if rng.Float64() < 0.38 {
			return TypeCrater
		}
	case "dust_sea":
		if rng.Float64() < 0.42 {
			return TypeDustField
		}
		if rng.Float64() < 0.28 {
			return TypeColdSink
		}
	case "cold_front":
		if r >= MapRows-2 {
			return TypeColdSink
		}
	}
	return rollTerrain(rng)
}

func rollTerrain(rng *rand.Rand) HexType {
	x := rng.Float64()
	switch {
	case x < 0.40:
		return TypeRegolith
	case x < 0.55:
		return TypeSolarPlateau
	case x < 0.68:
		return TypeCrater
	case x < 0.80:
		return TypeRidge
	case x < 0.90:
		return TypeDustField
	default:
		return TypeColdSink
	}
}

func (s *GameState) rollContracts(rng *rand.Rand, westBase, eastBase string, ids []string) []Contract {
	order := rng.Perm(len(cargoPool))
	used := map[string]bool{westBase: true, eastBase: true}
	n := 5 + rng.Intn(3)
	if n > len(order) {
		n = len(order)
	}
	out := make([]Contract, 0, n)
	for i := 0; i < n; i++ {
		spec := cargoPool[order[i]]
		pick := s.pickUniqueHex(rng, ids, used, false)
		used[pick] = true
		drop := s.farBaseID()
		ph := s.Map[pick]
		risk := hexRisk(ph)
		urgency := "low"
		deadline := 0.0
		if spec.cargo == CargoMedSeeds {
			deadline = 70 + rng.Float64()*28
			urgency = "high"
		} else if risk == "high" {
			urgency = "high"
		} else if spec.weight == WeightClassHeavy {
			urgency = "medium"
		}
		out = append(out, Contract{
			ID:          fmt.Sprintf("c%d", i),
			Title:       spec.title,
			Cargo:       spec.cargo,
			Weight:      spec.weight,
			Pickup:      pick,
			Dropoff:     drop,
			ColonyValue: spec.colony,
			EarthValue:  spec.earth,
			Reward:      spec.colony + spec.earth,
			Risk:        risk,
			Urgency:     urgency,
			Deadline:    deadline,
			Status:      ContractQueued,
		})
	}
	return out
}

func (s *GameState) boostDeliverableCeiling() {
	sum := 0
	idx := make([]int, 0, len(s.Contracts))
	for i, c := range s.Contracts {
		sum += c.ColonyValue
		idx = append(idx, i)
	}
	need := ColonyWinThreshold
	for sum < need && len(idx) > 0 {
		grew := false
		for _, i := range idx {
			if sum >= need {
				break
			}
			s.Contracts[i].ColonyValue += 2
			s.Contracts[i].Reward = s.Contracts[i].ColonyValue + s.Contracts[i].EarthValue
			sum += 2
			grew = true
		}
		if !grew {
			break
		}
	}
}

func (s *GameState) pickUniqueHex(rng *rand.Rand, ids []string, used map[string]bool, farthest bool) string {
	from := Axial{Q: s.Rovers[0].Q, R: s.Rovers[0].R}
	best := ""
	bestD := -1
	perm := rng.Perm(len(ids))
	fallback := ""
	for _, i := range perm {
		id := ids[i]
		if used[id] {
			continue
		}
		h := s.Map[id]
		if fallback == "" {
			fallback = id
		}
		if h.Type == TypeBase {
			continue
		}
		if farthest {
			ax, err := ParseHexID(id)
			if err != nil {
				continue
			}
			d := CubeDistance(from, ax)
			if d > bestD {
				bestD = d
				best = id
			}
			continue
		}
		return id
	}
	if farthest && best != "" {
		return best
	}
	if fallback != "" {
		return fallback
	}
	return ids[0]
}
