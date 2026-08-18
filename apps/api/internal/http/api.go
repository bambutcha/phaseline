package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"phaseline/internal/db"
	"phaseline/internal/game"
	"phaseline/internal/sim"
)

type Server struct {
	Hub   *game.Hub
	Pool  *pgxpool.Pool
	Query *db.Queries
}

type createGameRequest struct {
	Seed  string `json:"seed"`
	Rover string `json:"rover"`
}

type createGameResponse struct {
	ID      uuid.UUID `json:"id"`
	Verdict string    `json:"verdict,omitempty"`
	Share   string    `json:"shareUrl,omitempty"`
	sim.Snapshot
}

func New(hub *game.Hub, pool *pgxpool.Pool) *Server {
	s := &Server{Hub: hub, Pool: pool, Query: db.New(pool)}
	hub.OnFinish = s.persistFinish
	return s
}

func (s *Server) Routes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	v1.POST("/games", s.createGame)
	v1.GET("/games/:id", s.getGame)
	v1.POST("/games/:id/contracts/:cid/accept", s.accept)
	v1.POST("/games/:id/route", s.setRoute)
	v1.POST("/games/:id/deploy", s.deploy)
	v1.POST("/games/:id/reroute", s.reroute)
	v1.POST("/games/:id/autonomy", s.autonomy)
	v1.POST("/games/:id/goto", s.gotoHex)
	v1.GET("/games/:id/blackbox", s.blackbox)
	v1.GET("/seeds/:seed", s.seedPreview)
	r.GET("/ws/game/:id", s.wsGame)
}

func writeErr(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
}

func (s *Server) createGame(c *gin.Context) {
	var req createGameRequest
	_ = c.ShouldBindJSON(&req)
	rover := sim.RoverSwift
	if req.Rover == string(sim.RoverHauler) {
		rover = sim.RoverHauler
	}
	st := sim.NewGame(req.Seed, rover)
	if ghost, err := s.Query.GetGhostBySeed(c.Request.Context(), st.Seed); err == nil {
		var gr sim.GhostReplay
		if json.Unmarshal(ghost.ReplayJson, &gr) == nil && len(gr.Points) > 0 {
			st.Ghost = &gr
		}
	}
	snap := st.Snapshot()
	mapJSON, _ := json.Marshal(snap.Map)
	termJSON, _ := json.Marshal(snap.Map.Terminator)

	row, err := s.Query.CreateGame(c.Request.Context(), db.CreateGameParams{
		Seed:           st.Seed,
		Status:         string(st.Status),
		MapJson:        mapJSON,
		Crisis:         st.CrisisKind,
		TerminatorJson: termJSON,
		RoverType:      string(st.R().Type),
	})
	if err != nil {
		writeErr(c, http.StatusInternalServerError, "invalid", err.Error())
		return
	}
	for _, ct := range st.Contracts {
		payload, _ := json.Marshal(ct)
		if _, err := s.Query.InsertContract(c.Request.Context(), db.InsertContractParams{
			GameID:      row.ID,
			PayloadJson: payload,
			Status:      string(ct.Status),
		}); err != nil {
			writeErr(c, http.StatusInternalServerError, "invalid", err.Error())
			return
		}
	}
	s.Hub.Put(&game.Runtime{ID: row.ID, State: st})
	c.JSON(http.StatusCreated, createGameResponse{ID: row.ID, Snapshot: snap})
}

func (s *Server) runtimeID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErr(c, http.StatusBadRequest, "invalid", "bad game id")
		return uuid.Nil, false
	}
	return id, true
}

func (s *Server) getGame(c *gin.Context) {
	id, ok := s.runtimeID(c)
	if !ok {
		return
	}
	rt, err := s.Hub.Mutate(id, func(*sim.GameState) error { return nil })
	if err != nil {
		writeErr(c, http.StatusNotFound, "not_found", "game not found")
		return
	}
	c.JSON(http.StatusOK, createGameResponse{ID: rt.ID, Snapshot: rt.State.Snapshot()})
}

func (s *Server) accept(c *gin.Context) {
	id, ok := s.runtimeID(c)
	if !ok {
		return
	}
	rt, err := s.Hub.Mutate(id, func(st *sim.GameState) error {
		return st.Accept(c.Param("cid"))
	})
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			writeErr(c, http.StatusNotFound, "not_found", "game not found")
			return
		}
		mapSimErr(c, err)
		return
	}
	c.JSON(http.StatusOK, createGameResponse{ID: rt.ID, Snapshot: rt.State.Snapshot()})
}

type routeRequest struct {
	HexPath []string `json:"hexPath"`
}

func (s *Server) setRoute(c *gin.Context) {
	id, ok := s.runtimeID(c)
	if !ok {
		return
	}
	var req routeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, http.StatusBadRequest, "invalid", "hexPath required")
		return
	}
	var pred sim.Prediction
	rt, err := s.Hub.Mutate(id, func(st *sim.GameState) error {
		if err := st.SetRoute(req.HexPath); err != nil {
			return err
		}
		pred = st.Predict(append([]sim.Axial(nil), st.R().Path...))
		return nil
	})
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			writeErr(c, http.StatusNotFound, "not_found", "game not found")
			return
		}
		mapSimErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"game":       createGameResponse{ID: rt.ID, Snapshot: rt.State.Snapshot()},
		"prediction": pred,
	})
}

func (s *Server) deploy(c *gin.Context) {
	id, ok := s.runtimeID(c)
	if !ok {
		return
	}
	rt, err := s.Hub.Mutate(id, func(st *sim.GameState) error {
		return st.Deploy()
	})
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			writeErr(c, http.StatusNotFound, "not_found", "game not found")
			return
		}
		mapSimErr(c, err)
		return
	}
	c.JSON(http.StatusOK, createGameResponse{ID: rt.ID, Snapshot: rt.State.Snapshot()})
}

func (s *Server) reroute(c *gin.Context) {
	id, ok := s.runtimeID(c)
	if !ok {
		return
	}
	var req routeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, http.StatusBadRequest, "invalid", "hexPath required")
		return
	}
	rt, err := s.Hub.Mutate(id, func(st *sim.GameState) error {
		return st.Reroute(req.HexPath)
	})
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			writeErr(c, http.StatusNotFound, "not_found", "game not found")
			return
		}
		mapSimErr(c, err)
		return
	}
	c.JSON(http.StatusOK, createGameResponse{ID: rt.ID, Snapshot: rt.State.Snapshot()})
}

func (s *Server) autonomy(c *gin.Context) {
	id, ok := s.runtimeID(c)
	if !ok {
		return
	}
	var applied bool
	var reason string
	rt, err := s.Hub.Mutate(id, func(st *sim.GameState) error {
		applied, reason = st.Autonomy()
		return nil
	})
	if err != nil {
		writeErr(c, http.StatusNotFound, "not_found", "game not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"applied": applied,
		"reason":  reason,
		"game":    createGameResponse{ID: rt.ID, Snapshot: rt.State.Snapshot()},
	})
}

func (s *Server) seedPreview(c *gin.Context) {
	st := sim.NewGame(c.Param("seed"), sim.RoverSwift)
	snap := st.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"seed": snap.Seed,
		"map": gin.H{
			"hexCount":   len(snap.Map.Hexes),
			"terminator": snap.Map.Terminator,
		},
	})
}

type gotoRequest struct {
	Hex string `json:"hex"`
}

func (s *Server) gotoHex(c *gin.Context) {
	id, ok := s.runtimeID(c)
	if !ok {
		return
	}
	var req gotoRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Hex == "" {
		writeErr(c, http.StatusBadRequest, "invalid", "hex required")
		return
	}
	var pred sim.Prediction
	rt, err := s.Hub.Mutate(id, func(st *sim.GameState) error {
		if err := st.GoTo(req.Hex); err != nil {
			return err
		}
		pred = st.Predict(append([]sim.Axial(nil), st.R().Path...))
		return nil
	})
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			writeErr(c, http.StatusNotFound, "not_found", "game not found")
			return
		}
		mapSimErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"game":       createGameResponse{ID: rt.ID, Snapshot: rt.State.Snapshot()},
		"prediction": pred,
	})
}

func (s *Server) blackbox(c *gin.Context) {
	id, ok := s.runtimeID(c)
	if !ok {
		return
	}
	rt, err := s.Hub.Mutate(id, func(*sim.GameState) error { return nil })
	if err != nil {
		writeErr(c, http.StatusNotFound, "not_found", "game not found")
		return
	}
	if rt.State.Status != sim.StatusFinished {
		writeErr(c, http.StatusConflict, "conflict", "game not finished")
		return
	}
	c.JSON(http.StatusOK, createGameResponse{
		ID:       rt.ID,
		Snapshot: rt.State.Snapshot(),
		Verdict:  sim.VerdictText(rt.State.Outcome, rt.State.Seed),
		Share:    "/s/" + rt.State.Seed,
	})
}

func (s *Server) persistFinish(id uuid.UUID, st *sim.GameState) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.Query.UpdateGameStatus(ctx, db.UpdateGameStatusParams{
		ID:          id,
		Status:      string(st.Status),
		ColonyScore: int32(st.ColonyScore),
		EarthScore:  int32(st.EarthScore),
		FinishedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		slog.Error("persist finish", "err", err)
		return
	}
	for _, ev := range st.Events {
		payload, _ := json.Marshal(ev.Payload)
		if _, err := s.Query.InsertEvent(ctx, db.InsertEventParams{
			GameID:      id,
			T:           ev.T,
			Kind:        ev.Kind,
			PayloadJson: payload,
		}); err != nil {
			slog.Error("persist event", "err", err)
			return
		}
	}
	replay := sim.BuildGhostReplay(st)
	replayJSON, _ := json.Marshal(replay)
	if _, err := s.Query.UpsertGhost(ctx, db.UpsertGhostParams{
		Seed:        st.Seed,
		GameID:      id,
		ColonyScore: int32(st.ColonyScore),
		EarthScore:  int32(st.EarthScore),
		ReplayJson:  replayJSON,
	}); err != nil {
		slog.Error("persist ghost", "err", err)
	}
}

func mapSimErr(c *gin.Context, err error) {
	switch {
	case sim.IsNotFound(err):
		writeErr(c, http.StatusNotFound, "not_found", err.Error())
	case sim.IsConflict(err):
		writeErr(c, http.StatusConflict, "conflict", err.Error())
	case sim.IsGameOver(err):
		writeErr(c, http.StatusConflict, "game_over", err.Error())
	default:
		writeErr(c, http.StatusBadRequest, "invalid", err.Error())
	}
}
