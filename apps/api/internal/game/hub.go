package game

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"phaseline/internal/sim"
)

var ErrNotFound = errors.New("not_found")

type Runtime struct {
	ID    uuid.UUID
	State *sim.GameState
}

type Hub struct {
	mu    sync.Mutex
	games map[uuid.UUID]*Runtime
}

func NewHub() *Hub {
	return &Hub{games: make(map[uuid.UUID]*Runtime)}
}

func (h *Hub) Put(rt *Runtime) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.games[rt.ID] = rt
}

func (h *Hub) Get(id uuid.UUID) (*Runtime, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rt, ok := h.games[id]
	return rt, ok
}

func (h *Hub) Mutate(id uuid.UUID, fn func(*sim.GameState) error) (*Runtime, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rt, ok := h.games[id]
	if !ok {
		return nil, ErrNotFound
	}
	if err := fn(rt.State); err != nil {
		return rt, err
	}
	return rt, nil
}
