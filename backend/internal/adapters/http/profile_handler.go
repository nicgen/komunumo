package http

import (
	"encoding/json"
	"net/http"

	"komunumo/backend/internal/adapters/http/middleware"
	"komunumo/backend/internal/application/profile"
)

type ProfileHandler struct {
	getSvc    *profile.GetProfileService
	updateSvc *profile.UpdateProfileService
	uploadSvc *profile.UploadAvatarService
}

func NewProfileHandler(
	getSvc *profile.GetProfileService,
	updateSvc *profile.UpdateProfileService,
	uploadSvc *profile.UploadAvatarService,
) *ProfileHandler {
	return &ProfileHandler{
		getSvc:    getSvc,
		updateSvc: updateSvc,
		uploadSvc: uploadSvc,
	}
}

func sessionIDFromContext(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(middleware.SessionIDKey).(string)
	return v, ok && v != ""
}

func (h *ProfileHandler) HandleGetMyProfile(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := sessionIDFromContext(r)
	if !ok {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	out, err := h.getSvc.GetMyProfile(r.Context(), sessionID)
	if err != nil {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *ProfileHandler) HandleUpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := sessionIDFromContext(r)
	if !ok {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	var in profile.UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.updateSvc.UpdateProfile(r.Context(), sessionID, clientIP(r), in); err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProfileHandler) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := sessionIDFromContext(r)
	if !ok {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
	file, header, err := r.FormFile("avatar")
	if err != nil {
		jsonError(w, "invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	path, err := h.uploadSvc.UploadAvatar(r.Context(), sessionID, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"avatar_path": path})
}
