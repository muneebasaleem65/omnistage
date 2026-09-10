package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/muneebasaleem65/omnistage/internal/config"
	"github.com/muneebasaleem65/omnistage/internal/http/handlers/auth"
	"github.com/muneebasaleem65/omnistage/internal/http/middleware"
	"github.com/muneebasaleem65/omnistage/internal/storage/postgres"
)

func main() {
	//load config
	cfg := config.MustLoad()

	//database setup

	storage, err := postgres.New(cfg.StoragePath)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	slog.Info("storage connected")

	authMW := middleware.Auth(cfg.JWTSecret)
	//setup router
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/auth/register", auth.Register(storage))
	mux.HandleFunc("POST /api/v1/auth/login", auth.Login(storage, cfg.JWTSecret))
	mux.Handle("GET /api/v1/auth/me", authMW(auth.Me(storage)))
	//setup server
	server := http.Server{
		Addr:    cfg.Address,
		Handler: mux,
	}

	slog.Info("server started", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("failed to start server")
		}
	}()

	<-done

	slog.Info("shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = server.Shutdown(ctx)

	if err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("server shutdown successfully")
}
