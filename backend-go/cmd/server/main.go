package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"plat-detection-system/backend-go/internal/config"
	"plat-detection-system/backend-go/internal/router"
)

func main() {
	cfg := config.Load()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router.New(cfg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("listening", "addr", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err.Error())
		os.Exit(1)
	}
}
