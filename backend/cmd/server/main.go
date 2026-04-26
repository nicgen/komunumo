package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"komunumo/backend/internal/adapters/clock"
	"komunumo/backend/internal/adapters/db"
	"komunumo/backend/internal/adapters/http/middleware"
	"komunumo/backend/internal/adapters/log"
	"komunumo/backend/internal/adapters/password"
	"komunumo/backend/internal/adapters/ratelimit"
	"komunumo/backend/internal/adapters/tokengen"
)

func main() {
	logger := log.New(os.Stdout, slog.LevelInfo)
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	dsn := envOr("KOMUNUMO_SQLITE_DSN", "./komunumo.db")
	addr := envOr("KOMUNUMO_HTTP_ADDR", ":8080")

	conn, err := db.Open(dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Adapter wiring. Repositories + use cases will be added in later tasks
	// (T048+). For now we keep the composition root minimal: it must compile,
	// boot, and serve a 501 on every /api/v1/auth/* route.
	_ = clock.New()
	_ = password.New()
	_ = tokengen.New()
	_ = ratelimit.New(60, time.Minute)

	r := chi.NewRouter()
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CSRF)

	r.Route("/api/v1/auth", func(r chi.Router) {
		notImpl := func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "not implemented", http.StatusNotImplemented)
		}
		r.Post("/register", notImpl)
		r.Post("/verify-email", notImpl)
		r.Post("/resend-verification", notImpl)
		r.Post("/login", notImpl)
		r.Post("/logout", notImpl)
		r.Get("/me", notImpl)
		r.Post("/password-reset/request", notImpl)
		r.Post("/password-reset/confirm", notImpl)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("http server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
