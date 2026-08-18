package sim

import "errors"

var (
	errGameOver     = errors.New("game_over")
	errInvalidRoute = errors.New("invalid")
	errConflict     = errors.New("conflict")
	errNotFound     = errors.New("not_found")
)

func IsGameOver(err error) bool { return errors.Is(err, errGameOver) }
func IsInvalid(err error) bool  { return errors.Is(err, errInvalidRoute) }
func IsConflict(err error) bool { return errors.Is(err, errConflict) }
func IsNotFound(err error) bool { return errors.Is(err, errNotFound) }

func (s *GameState) emit(kind string, payload map[string]any) {
	s.Events = append(s.Events, Event{T: round1(s.T), Kind: kind, Payload: payload})
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
