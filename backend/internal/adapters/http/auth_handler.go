package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/token"
)

const sessionCookieName = "session_id"
const sessionCookieMaxAge = int(30 * 24 * time.Hour / time.Second)

// AuthHandler wires application services to HTTP.
// Nil services indicate features not yet wired (returns 501).
type AuthHandler struct {
	register       *auth.RegisterService
	verify         *auth.VerifyEmailService
	resend         *auth.ResendVerificationService
	login          *auth.LoginService
	logout         *auth.LogoutService
	pwResetRequest *auth.PasswordResetRequestService
	pwResetConfirm *auth.PasswordResetConfirmService
	me             *auth.MeService
}

func NewAuthHandler(
	register *auth.RegisterService,
	verify *auth.VerifyEmailService,
	resend *auth.ResendVerificationService,
	login *auth.LoginService,
	logout *auth.LogoutService,
	pwResetRequest *auth.PasswordResetRequestService,
	pwResetConfirm *auth.PasswordResetConfirmService,
	me *auth.MeService,
) *AuthHandler {
	return &AuthHandler{
		register: register, verify: verify, resend: resend,
		login: login, logout: logout,
		pwResetRequest: pwResetRequest, pwResetConfirm: pwResetConfirm,
		me: me,
	}
}

// --- Register ---

type registerRequest struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"` // "YYYY-MM-DD"
	Password    string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in auth.RegisterInput
	isJSON := isJSONRequest(r)

	if isJSON {
		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			jsonError(w, "invalid date_of_birth", http.StatusBadRequest)
			return
		}
		in = auth.RegisterInput{
			Email:       req.Email,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			DateOfBirth: dob,
			Password:    req.Password,
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		dob, err := time.Parse("2006-01-02", r.FormValue("date_of_birth"))
		if err != nil {
			http.Error(w, "invalid date_of_birth", http.StatusBadRequest)
			return
		}
		in = auth.RegisterInput{
			Email:       r.FormValue("email"),
			FirstName:   r.FormValue("first_name"),
			LastName:    r.FormValue("last_name"),
			DateOfBirth: dob,
			Password:    r.FormValue("password"),
		}
	}

	ip := clientIP(r)
	if err := h.register.Register(r.Context(), ip, in); err != nil {
		handleRegisterError(w, r, err, isJSON)
		return
	}

	if isJSON {
		w.WriteHeader(http.StatusCreated)
	} else {
		http.Redirect(w, r, "/verify-email/sent", http.StatusSeeOther)
	}
}

func handleRegisterError(w http.ResponseWriter, _ *http.Request, err error, isJSON bool) {
	var status int
	var msg string

	switch {
	case errors.Is(err, account.ErrAgeBelow16):
		status, msg = http.StatusBadRequest, "vous devez avoir au moins 16 ans"
	case errors.Is(err, account.ErrEmailMalformed):
		status, msg = http.StatusBadRequest, "adresse email invalide"
	case errors.Is(err, account.ErrPasswordTooShort), errors.Is(err, account.ErrPasswordTooWeak):
		status, msg = http.StatusBadRequest, "mot de passe trop faible"
	default:
		status, msg = http.StatusInternalServerError, "erreur interne"
	}

	if isJSON {
		jsonError(w, msg, status)
	} else {
		http.Error(w, msg, status)
	}
}

// --- VerifyEmail ---

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var rawToken string
	isJSON := isJSONRequest(r)

	if isJSON {
		var req verifyEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		rawToken = req.Token
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		rawToken = r.FormValue("token")
	}

	if rawToken == "" {
		if isJSON {
			jsonError(w, "token required", http.StatusBadRequest)
		} else {
			http.Error(w, "token required", http.StatusBadRequest)
		}
		return
	}

	err := h.verify.VerifyEmail(r.Context(), auth.VerifyEmailInput{RawToken: rawToken})
	if err != nil {
		handleVerifyError(w, r, err, isJSON)
		return
	}

	if isJSON {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/login?verified=1", http.StatusSeeOther)
	}
}

func handleVerifyError(w http.ResponseWriter, _ *http.Request, err error, isJSON bool) {
	var status int
	var msg string

	switch {
	case errors.Is(err, token.ErrTokenExpired):
		status, msg = http.StatusGone, "lien expiré"
	case errors.Is(err, token.ErrTokenNotFound), errors.Is(err, token.ErrTokenAlreadyConsumed):
		status, msg = http.StatusBadRequest, "lien invalide"
	default:
		status, msg = http.StatusInternalServerError, "erreur interne"
	}

	if isJSON {
		jsonError(w, msg, status)
	} else {
		http.Error(w, msg, status)
	}
}

// --- ResendVerification ---

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	if isJSON := isJSONRequest(r); isJSON {
		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		input := auth.ResendVerificationInput{Email: req.Email, IP: clientIP(r)}
		if err := h.resend.Resend(r.Context(), input); err != nil {
			if errors.Is(err, auth.ErrRateLimited) {
				jsonError(w, "trop de tentatives", http.StatusTooManyRequests)
				return
			}
			jsonError(w, "erreur interne", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		input := auth.ResendVerificationInput{Email: r.FormValue("email"), IP: clientIP(r)}
		if err := h.resend.Resend(r.Context(), input); err != nil {
			if errors.Is(err, auth.ErrRateLimited) {
				http.Error(w, "trop de tentatives", http.StatusTooManyRequests)
				return
			}
			http.Error(w, "erreur interne", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/verify-email/sent", http.StatusSeeOther)
	}
}

// --- Login ---

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.login == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	var email, password string
	isJSON := isJSONRequest(r)

	if isJSON {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		email, password = req.Email, req.Password
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		email = r.FormValue("email")
		password = r.FormValue("password")
	}

	out, err := h.login.Login(r.Context(), auth.LoginInput{
		Email:     email,
		Password:  password,
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		handleLoginError(w, r, err, isJSON)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    out.SessionID,
		MaxAge:   sessionCookieMaxAge,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	if isJSON {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func handleLoginError(w http.ResponseWriter, _ *http.Request, err error, isJSON bool) {
	var status int
	var msg string

	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		status, msg = http.StatusUnauthorized, "identifiants incorrects"
	case errors.Is(err, account.ErrAccountNotVerified):
		status, msg = http.StatusForbidden, "email non vérifié"
	case errors.Is(err, account.ErrAccountDisabled):
		status, msg = http.StatusForbidden, "compte désactivé"
	case errors.Is(err, auth.ErrRateLimited):
		status, msg = http.StatusTooManyRequests, "trop de tentatives"
	default:
		status, msg = http.StatusInternalServerError, "erreur interne"
	}

	if isJSON {
		jsonError(w, msg, status)
	} else {
		http.Error(w, msg, status)
	}
}

// --- Logout ---

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if h.logout == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = h.logout.Logout(r.Context(), cookie.Value)
	}

	// Clear the cookie regardless of whether session existed.
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	isJSON := isJSONRequest(r)
	if isJSON {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// --- PasswordResetRequest ---

func (h *AuthHandler) PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	if h.pwResetRequest == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	var email string
	isJSON := isJSONRequest(r)

	if isJSON {
		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		email = req.Email
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		email = r.FormValue("email")
	}

	err := h.pwResetRequest.Request(r.Context(), auth.PasswordResetRequestInput{
		Email: email,
		IP:    clientIP(r),
	})
	if err != nil {
		if errors.Is(err, auth.ErrRateLimited) {
			if isJSON {
				jsonError(w, "trop de tentatives", http.StatusTooManyRequests)
			} else {
				http.Error(w, "trop de tentatives", http.StatusTooManyRequests)
			}
			return
		}
		if isJSON {
			jsonError(w, "erreur interne", http.StatusInternalServerError)
		} else {
			http.Error(w, "erreur interne", http.StatusInternalServerError)
		}
		return
	}

	// Always 200 regardless of whether the email exists (anti-enumeration).
	if isJSON {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/reset-password/sent", http.StatusSeeOther)
	}
}

// --- PasswordResetConfirm ---

func (h *AuthHandler) PasswordResetConfirm(w http.ResponseWriter, r *http.Request) {
	if h.pwResetConfirm == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	var rawToken, newPassword string
	isJSON := isJSONRequest(r)

	if isJSON {
		var req struct {
			Token       string `json:"token"`
			NewPassword string `json:"new_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		rawToken, newPassword = req.Token, req.NewPassword
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		rawToken = r.FormValue("token")
		newPassword = r.FormValue("new_password")
	}

	err := h.pwResetConfirm.Confirm(r.Context(), auth.PasswordResetConfirmInput{
		RawToken:    rawToken,
		NewPassword: newPassword,
	})
	if err != nil {
		handlePasswordResetConfirmError(w, r, err, isJSON)
		return
	}

	if isJSON {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Redirect(w, r, "/login?password_reset=1", http.StatusSeeOther)
	}
}

func handlePasswordResetConfirmError(w http.ResponseWriter, _ *http.Request, err error, isJSON bool) {
	var status int
	var msg string

	switch {
	case errors.Is(err, token.ErrTokenNotFound), errors.Is(err, token.ErrTokenExpired), errors.Is(err, token.ErrTokenAlreadyConsumed):
		status, msg = http.StatusBadRequest, "lien invalide ou expiré"
	case errors.Is(err, account.ErrPasswordTooShort), errors.Is(err, account.ErrPasswordTooWeak):
		status, msg = http.StatusBadRequest, "mot de passe trop faible"
	default:
		status, msg = http.StatusInternalServerError, "erreur interne"
	}

	if isJSON {
		jsonError(w, msg, status)
	} else {
		http.Error(w, msg, status)
	}
}

// --- Me ---

type meResponse struct {
	AccountID string `json:"account_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Status    string `json:"status"`
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if h.me == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	out, err := h.me.Me(r.Context(), cookie.Value)
	if err != nil {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(meResponse{
		AccountID: out.AccountID,
		Email:     out.Email,
		FirstName: out.FirstName,
		LastName:  out.LastName,
		Status:    string(out.Status),
	})
}

// --- helpers ---

func isJSONRequest(r *http.Request) bool {
	return r.Header.Get("Content-Type") == "application/json"
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	return r.RemoteAddr
}
