package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"phaseline/internal/sim"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(*http.Request) bool { return true },
}

type wsIn struct {
	Type       string   `json:"type"`
	ContractID string   `json:"contractId"`
	HexPath    []string `json:"hexPath"`
	HexID      string   `json:"hexId"`
	Rover      string   `json:"rover"`
}

func (s *Server) wsGame(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErr(c, http.StatusBadRequest, "invalid", "bad game id")
		return
	}
	subID, ch, snap, err := s.Hub.Subscribe(id)
	if err != nil {
		writeErr(c, http.StatusNotFound, "not_found", "game not found")
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.Hub.Unsubscribe(id, subID)
		return
	}
	defer func() {
		s.Hub.Unsubscribe(id, subID)
		conn.Close()
	}()

	_ = conn.WriteMessage(websocket.TextMessage, snap)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var msg wsIn
			if err := json.Unmarshal(raw, &msg); err != nil {
				continue
			}
			if _, err := s.Hub.Mutate(id, func(st *sim.GameState) error {
				return applyIntent(st, msg)
			}); err != nil {
				slog.Debug("ws intent", "err", err)
			}
		}
	}()

	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-done:
			return
		case payload, ok := <-ch:
			if !ok {
				return
			}
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ping.C:
			_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func applyIntent(st *sim.GameState, msg wsIn) error {
	if msg.Rover == string(sim.RoverHauler) {
		st.SelectRover(sim.RoverHauler)
	} else if msg.Rover == string(sim.RoverSwift) {
		st.SelectRover(sim.RoverSwift)
	}
	switch msg.Type {
	case "accept_contract":
		return st.Accept(msg.ContractID)
	case "dispatch":
		return st.Dispatch(msg.ContractID)
	case "set_route":
		return st.SetRoute(msg.HexPath)
	case "goto":
		return st.GoTo(msg.HexID)
	case "deploy":
		return st.Deploy()
	case "reroute":
		return st.Reroute(msg.HexPath)
	case "autonomy":
		_, _ = st.Autonomy()
		return nil
	case "select_rover":
		if msg.Rover == string(sim.RoverHauler) {
			st.SelectRover(sim.RoverHauler)
		} else {
			st.SelectRover(sim.RoverSwift)
		}
		return nil
	default:
		return sim.ErrInvalid
	}
}
