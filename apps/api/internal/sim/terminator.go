package sim

// Hex center is in shadow when the terminator has passed it.
// East: pos increases; shadow is q+0.5 < pos.
// West: pos decreases; shadow is q-0.5 > pos.
func (t Terminator) InShadow(q int) bool {
	c := float64(q)
	switch t.Direction {
	case DirWest:
		return c-0.5 > t.Pos
	default:
		return c+0.5 < t.Pos
	}
}

func (t Terminator) Advance(dt float64) Terminator {
	switch t.Direction {
	case DirWest:
		t.Pos -= t.Speed * dt
	default:
		t.Pos += t.Speed * dt
	}
	return t
}

// PhaseETA is seconds until this hex enters shadow, or 0 if already dark.
func (t Terminator) PhaseETA(q int) float64 {
	if t.Speed <= 0 {
		return 0
	}
	if t.InShadow(q) {
		return 0
	}
	c := float64(q)
	var dist float64
	switch t.Direction {
	case DirWest:
		dist = t.Pos - (c - 0.5)
	default:
		dist = (c + 0.5) - t.Pos
	}
	if dist < 0 {
		return 0
	}
	return dist / t.Speed
}

func (s *GameState) InShadow(q, r int) bool {
	return s.Terminator.InShadow(q)
}
