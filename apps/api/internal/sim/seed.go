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
	{CargoO2, "O₂-фильтры — Hab Block C", WeightClassMedium, 40, 10, 0},
	{CargoO2, "Скрубберы CO₂ — шлюз 2", WeightClassMedium, 35, 8, 0},
	{CargoCryo, "Криопробы — кратер Шиккард", WeightClassLight, 25, 30, 0},
	{CargoCrew, "Капсула экипажа — смена «Рассвет»", WeightClassMedium, 50, 5, 0},
	{CargoReactor, "Стержень реактора — ядро Alpha", WeightClassHeavy, 20, 35, 0},
	{CargoHelium3, "Helium-3 — контракт Земли", WeightClassHeavy, 5, 50, 0},
	{CargoMedSeeds, "Медсемена — оранжерея", WeightClassLight, 30, 15, 90},
	{CargoCommRelay, "Реле связи — пик хребта", WeightClassLight, 20, 20, 0},
}

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

	hexes := make(map[string]Hex, MapHexCount)
	var ids []string
	for r := 0; r < 3; r++ {
		for q := 0; q < 4; q++ {
			h := Hex{Q: q, R: r, Type: rollTerrain(rng)}
			hexes[h.ID()] = h
			ids = append(ids, h.ID())
		}
	}

	westBase := Hex{Q: 0, R: 1, Type: TypeBase}
	eastBase := Hex{Q: 3, R: 1, Type: TypeBase}
	hexes[westBase.ID()] = westBase
	hexes[eastBase.ID()] = eastBase

	start := westBase.Axial()
	if dir == DirWest {
		start = eastBase.Axial()
	}

	term := Terminator{Speed: TerminatorSpeed, Direction: dir}
	if dir == DirWest {
		term.Pos = 3.5
	} else {
		term.Pos = -0.5
	}

	crisis := crisisKinds[rng.Intn(len(crisisKinds))]
	jitter := rng.Float64()*20 - 10
	crisisAt := 0.4*GameDurationTargetSec + jitter

	g := &GameState{
		Seed:         seed,
		Status:       StatusLobby,
		Map:          hexes,
		Terminator:   term,
		Rover:        NewRover(rover, start.Q, start.R),
		CrisisKind:   crisis,
		CrisisAt:     crisisAt,
		AutonomyLeft: AutonomyCharges,
		FreeReroutes: RerouteFreeCount,
	}
	g.Contracts = g.rollContracts(rng, westBase.ID(), eastBase.ID(), ids)
	return g
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
	out := make([]Contract, 0, ContractsPerGame)
	for i := 0; i < ContractsPerGame && i < len(order); i++ {
		spec := cargoPool[order[i]]
		pick := ids[rng.Intn(len(ids))]
		drop := eastBase
		if s.Terminator.Direction == DirWest {
			drop = westBase
		}
		if rng.Float64() < 0.35 {
			drop = ids[rng.Intn(len(ids))]
		}
		if pick == drop {
			drop = eastBase
			if pick == drop {
				drop = westBase
			}
		}
		deadline := spec.deadline
		if spec.cargo == CargoMedSeeds {
			deadline = 80 + rng.Float64()*40
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
			Deadline:    deadline,
			Status:      ContractQueued,
		})
	}
	return out
}
