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

	"komunumo/backend/internal/adapters/clock"
	"komunumo/backend/internal/adapters/db"
	"komunumo/backend/internal/adapters/email"
	httpadapter "komunumo/backend/internal/adapters/http"
	"komunumo/backend/internal/adapters/log"
	"komunumo/backend/internal/adapters/password"
	"komunumo/backend/internal/adapters/ratelimit"
	"komunumo/backend/internal/adapters/tokengen"
	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/ports"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
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
	brevoKey := envOr("KOMUNUMO_BREVO_API_KEY", "test-key-noop")
	appBaseURL := envOr("KOMUNUMO_APP_BASE_URL", "http://localhost:3000")
	fromEmail := envOr("KOMUNUMO_BREVO_FROM_EMAIL", "noreply@komunumo.fr")
	fromName := envOr("KOMUNUMO_BREVO_FROM_NAME", "Komunumo")

	conn, err := db.Open(dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	clk := clock.New()
	hasher := password.New()
	tokenGen := tokengen.New()
	rl := ratelimit.New(60, time.Minute)
	uow := db.NewUnitOfWork(conn)

	accounts := db.NewAccountRepository(conn)
	tokens := db.NewTokenRepository(conn)
	auditRepo := db.NewAuditRepository(conn)

	var emailSender ports.EmailSender
	if brevoKey == "" || brevoKey == "test-key-noop" || os.Getenv("NODE_ENV") == "development" {
		logger.Info("using log email sender")
		emailSender = email.NewLogSender(appBaseURL)
	} else {
		logger.Info("using brevo email sender")
		emailSender = email.NewBrevoSender(email.BrevoConfig{
			APIKey:     brevoKey,
			FromEmail:  fromEmail,
			FromName:   fromName,
			AppBaseURL: appBaseURL,
		})
	}

	registerSvc := auth.NewRegisterService(accounts, tokens, auditRepo, emailSender, hasher, tokenGen, clk, rl, uow)
	verifySvc := auth.NewVerifyEmailService(accounts, tokens, auditRepo, tokenGen, clk, uow)
	resendSvc := auth.NewResendVerificationService(accounts, tokens, auditRepo, emailSender, tokenGen, clk, rl, uow)
	sessionRepo := db.NewSessionRepository(conn)
	loginSvc := auth.NewLoginService(accounts, sessionRepo, auditRepo, hasher, tokenGen, clk, rl, uow)
	logoutSvc := auth.NewLogoutService(sessionRepo, auditRepo, tokenGen, clk)
	pwResetReqSvc := auth.NewPasswordResetRequestService(accounts, tokens, auditRepo, emailSender, tokenGen, clk, rl, uow)
	pwResetConfSvc := auth.NewPasswordResetConfirmService(accounts, tokens, sessionRepo, auditRepo, emailSender, hasher, tokenGen, clk, uow)
	meSvc := auth.NewMeService(sessionRepo, accounts, clk)

	authHandler := httpadapter.NewAuthHandler(registerSvc, verifySvc, resendSvc, loginSvc, logoutSvc, pwResetReqSvc, pwResetConfSvc, meSvc)
	router := httpadapter.NewRouter(authHandler)

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
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
