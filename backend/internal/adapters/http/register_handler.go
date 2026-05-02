package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"komunumo/backend/internal/application/auth"
	"komunumo/backend/internal/domain/account"
	"komunumo/backend/internal/domain/member"
)

type RegisterHandler struct {
	registerMember *auth.RegisterMemberService
}

func NewRegisterHandler(registerMember *auth.RegisterMemberService) *RegisterHandler {
	return &RegisterHandler{
		registerMember: registerMember,
	}
}

func (h *RegisterHandler) HandleRegisterMember(w http.ResponseWriter, r *http.Request) {
	if h.registerMember == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}

	var in auth.RegisterMemberInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.registerMember.RegisterMember(r.Context(), clientIP(r), in)
	if err != nil {
		handleRegisterMemberError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleRegisterMemberError(w http.ResponseWriter, _ *http.Request, err error) {
	var status int
	var msg string

	switch {
	case errors.Is(err, account.ErrEmailTaken):
		status, msg = http.StatusConflict, "email déjà utilisé"
	case errors.Is(err, member.ErrTooYoung):
		status, msg = http.StatusUnprocessableEntity, "vous devez avoir au moins 18 ans"
	case errors.Is(err, account.ErrPasswordTooShort), errors.Is(err, account.ErrPasswordTooWeak):
		status, msg = http.StatusBadRequest, "mot de passe trop faible"
	case errors.Is(err, auth.ErrRateLimited):
		status, msg = http.StatusTooManyRequests, "trop de tentatives"
	default:
		status, msg = http.StatusInternalServerError, "erreur interne"
	}

	jsonError(w, msg, status)
}
