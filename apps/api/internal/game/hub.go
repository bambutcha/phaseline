package game

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"phaseline/internal/sim"
)

var ErrNotFound = errors.New("not_found")

type Envelope struct {
	Type string    `json:"type"`
	ID   uuid.UUID `json:"id"`
	sim.Snapshot
	DeltaEvents []sim.Event `json:"deltaEvents,omitempty"`
}

type Runtime struct {
	ID       uuid.UUID
	State    *sim.GameState
	ticking  bool
	finished bool
	subs     map[int]chan []byte
	nextSub  int
}

type Hub struct {
	mu       sync.Mutex
	games    map[uuid.UUID]*Runtime
	OnFinish func(id uuid.UUID, st *sim.GameState)
}

func NewHub() *Hub {
	return &Hub{games: make(map[uuid.UUID]*Runtime)}
}

func (h *Hub) Put(rt *Runtime) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if rt.subs == nil {
		rt.subs = make(map[int]chan []byte)
	}
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
	rt, ok := h.games[id]
	if !ok {
		h.mu.Unlock()
		return nil, ErrNotFound
	}
	if err := fn(rt.State); err != nil {
		h.mu.Unlock()
		return rt, err
	}
	start := false
	if rt.State.Status == sim.StatusActive && !rt.ticking {
		rt.ticking = true
		start = true
	}
	h.broadcastLocked(rt, "snapshot", nil)
	h.mu.Unlock()
	if start {
		go h.tickLoop(id)
	}
	return rt, nil
}

func (h *Hub) Subscribe(id uuid.UUID) (int, <-chan []byte, []byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rt, ok := h.games[id]
	if !ok {
		return 0, nil, nil, ErrNotFound
	}
	if rt.subs == nil {
		rt.subs = make(map[int]chan []byte)
	}
	rt.nextSub++
	sid := rt.nextSub
	ch := make(chan []byte, 32)
	rt.subs[sid] = ch
	start := false
	if rt.State.Status == sim.StatusActive && !rt.ticking {
		rt.ticking = true
		start = true
	}
	snap := marshalEnv(rt, "snapshot", nil)
	if start {
		go h.tickLoop(id)
	}
	return sid, ch, snap, nil
}

func (h *Hub) Unsubscribe(id uuid.UUID, subID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rt, ok := h.games[id]
	if !ok {
		return
	}
	ch, ok := rt.subs[subID]
	if !ok {
		return
	}
	delete(rt.subs, subID)
	close(ch)
}

func (h *Hub) tickLoop(id uuid.UUID) {
	ticker := time.NewTicker(time.Second / sim.TickHz)
	defer ticker.Stop()
	h.tickOnce(id)
	for range ticker.C {
		if !h.tickOnce(id) {
			return
		}
	}
}

func (h *Hub) tickOnce(id uuid.UUID) bool {
	h.mu.Lock()
	rt, ok := h.games[id]
	if !ok {
		h.mu.Unlock()
		return false
	}
	if rt.State.Status != sim.StatusActive {
		done := rt.State.Status == sim.StatusFinished
		h.mu.Unlock()
		return !done
	}
	delta := rt.State.Tick(sim.TickDT)
	typ := "tick"
	if rt.State.Status == sim.StatusFinished {
		typ = "game_over"
	}
	h.broadcastLocked(rt, typ, delta)
	finish := rt.State.Status == sim.StatusFinished && !rt.finished
	var stCopy *sim.GameState
	if finish {
		rt.finished = true
		stCopy = rt.State.Clone()
	}
	alive := rt.State.Status != sim.StatusFinished
	cb := h.OnFinish
	h.mu.Unlock()
	if finish && cb != nil {
		cb(id, stCopy)
	}
	return alive
}

func (h *Hub) broadcastLocked(rt *Runtime, typ string, delta []sim.Event) {
	payload := marshalEnv(rt, typ, delta)
	for _, ch := range rt.subs {
		select {
		case ch <- payload:
		default:
		}
	}
}

func marshalEnv(rt *Runtime, typ string, delta []sim.Event) []byte {
	env := Envelope{Type: typ, ID: rt.ID, Snapshot: rt.State.Snapshot(), DeltaEvents: delta}
	b, _ := json.Marshal(env)
	return b
}
