package sim

type GhostPoint struct {
	T float64 `json:"t"`
	Q int     `json:"q"`
	R int     `json:"r"`
}

type GhostReplay struct {
	Points      []GhostPoint `json:"points"`
	ColonyScore int          `json:"colonyScore"`
	EarthScore  int          `json:"earthScore"`
}

func BuildGhostReplay(st *GameState) GhostReplay {
	points := make([]GhostPoint, 0, len(st.Events))
	for _, ev := range st.Events {
		if ev.Kind != "hex_entered" {
			continue
		}
		hid, _ := ev.Payload["hexId"].(string)
		ax, err := ParseHexID(hid)
		if err != nil {
			continue
		}
		points = append(points, GhostPoint{T: ev.T, Q: ax.Q, R: ax.R})
	}
	return GhostReplay{
		Points:      points,
		ColonyScore: st.ColonyScore,
		EarthScore:  st.EarthScore,
	}
}

func VerdictText(outcome Outcome, seed string) string {
	idx := int(seedHash(seed) % 3)
	switch outcome {
	case OutcomeSaved:
		texts := []string{
			"Вы спасли ядро колонии. Оставшиеся в тени молчат, но благодаря вам кто-то встретит рассвет.",
			"Идеальный triage. Колония выжила. Земля недовольна убытками — им не жить здесь.",
			"Решения были холодными, как вакуум. Колония дышит.",
		}
		return texts[idx]
	case OutcomePyrrhic:
		texts := []string{
			"Колония жива, экономика мертва. Следующий корабль Земли — через десять лет.",
			"Вы выбрали людей. Теперь есть рты, которые нечем кормить. Но они живы.",
			"Триумф над инструкцией. Зато есть кому рассказать, как вы провалили гелий-3.",
		}
		return texts[idx]
	default:
		texts := []string{
			"Линия сомкнулась. Сигнал ровера утонул в белом шуме. Колония молчит.",
			"Слишком много жадности, слишком мало воздуха. Тень забрала тех, кого вы не успели спасти.",
			"Вы пытались спасти всех. В итоге не спасся никто. Тень стёрла следы.",
		}
		return texts[idx]
	}
}

func seedHash(seed string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(seed); i++ {
		h ^= uint32(seed[i])
		h *= 16777619
	}
	return h
}
