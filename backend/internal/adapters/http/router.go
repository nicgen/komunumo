package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"komunumo/backend/internal/adapters/http/middleware"
)

func NewRouter(auth *AuthHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.SecurityHeaders)

	// Public auth endpoints — CSRF-exempt per D-009 (unauthenticated).
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/verify-email", auth.VerifyEmail)
		r.Post("/resend-verification", auth.ResendVerification)
		r.Post("/login", auth.Login)
		r.Post("/logout", auth.Logout)
		r.Get("/me", auth.Me)
		r.Post("/password-reset/request", auth.PasswordResetRequest)
		r.Post("/password-reset/confirm", auth.PasswordResetConfirm)
	})

	// Protected routes (Phase 4+) will go here with r.Use(middleware.CSRF).

	return r
}
