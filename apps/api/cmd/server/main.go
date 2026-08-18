package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"phaseline/internal/game"
	httpapi "phaseline/internal/http"
	"phaseline/internal/migrate"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://phaseline:phaseline@localhost:5432/phaseline?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("db pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()
	var pingErr error
	wait := time.Now().Add(30 * time.Second)
	for {
		pingErr = sqlDB.Ping()
		if pingErr == nil {
			break
		}
		if time.Now().After(wait) {
			slog.Error("db ping", "err", pingErr)
			os.Exit(1)
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err := migrate.Up(sqlDB); err != nil {
		slog.Error("migrate", "err", err)
		os.Exit(1)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ready", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	api := httpapi.New(game.NewHub(), pool)
	api.Routes(r)

	slog.Info("phaseline api listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
