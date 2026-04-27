package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"komunumo/backend/internal/adapters/http/middleware"
)

func NewRouter(auth *AuthHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CSRF)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", auth.Register)
		r.Post("/verify-email", auth.VerifyEmail)
		r.Post("/resend-verification", auth.ResendVerification)

		// Phase 4 (US2)
		notImpl := func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "not implemented", http.StatusNotImplemented)
		}
		r.Post("/login", notImpl)
		r.Post("/logout", notImpl)
		r.Get("/me", notImpl)
		r.Post("/password-reset/request", notImpl)
		r.Post("/password-reset/confirm", notImpl)
	})

	return r
}
