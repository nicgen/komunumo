package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"komunumo/backend/internal/adapters/http/middleware"
	"komunumo/backend/internal/ports"
)

func NewRouter(auth *AuthHandler, register *RegisterHandler, profile *ProfileHandler, sessions ports.SessionRepository, clk ports.Clock) http.Handler {
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

		r.Post("/register/member", register.HandleRegisterMember)
		r.Post("/register/association", register.HandleRegisterAssociation)
	})

	// Public profile endpoint (with optional auth for members_only visibility)
	r.With(middleware.OptionalAuth(sessions, clk)).Get("/api/v1/accounts/{accountId}/profile", profile.HandleGetPublicProfile)

	// Protected routes (US3+)
	r.Group(func(r chi.Router) {
		r.Use(middleware.CSRF)
		r.Use(middleware.Auth(sessions, clk))

		r.Get("/api/v1/me/profile", profile.HandleGetMyProfile)
		r.Patch("/api/v1/me/profile", profile.HandleUpdateMyProfile)
		r.Post("/api/v1/me/avatar", profile.HandleUploadAvatar)
	})

	return r
}
