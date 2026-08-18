package httpapi

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"phaseline/internal/game"
	"phaseline/internal/sim"
)

func TestWebSocketSnapshotAndGoto(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := game.NewHub()
	st := sim.NewGame("MCC-TEST", sim.RoverSwift)
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	hub.Put(&game.Runtime{ID: id, State: st})
	s := &Server{Hub: hub}
	r := gin.New()
	r.GET("/ws/game/:id", s.wsGame)
	ts := httptest.NewServer(r)
	defer ts.Close()

	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/game/" + id.String()
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := c.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var env game.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env.Type != "snapshot" || env.Status != sim.StatusLobby {
		t.Fatalf("got type=%s status=%s", env.Type, env.Status)
	}
	if err := c.WriteJSON(map[string]string{"type": "goto", "hexId": "3,1"}); err != nil {
		t.Fatal(err)
	}
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err = c.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Rover.Path) == 0 {
		t.Fatalf("expected path after goto: %+v", env.Rover)
	}
}
