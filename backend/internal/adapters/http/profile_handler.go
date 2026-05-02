package http

import (
	"encoding/json"
	"net/http"

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

func (h *ProfileHandler) HandleGetMyProfile(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	out, err := h.getSvc.GetMyProfile(r.Context(), cookie.Value)
	if err != nil {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *ProfileHandler) HandleUpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	var in profile.UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.updateSvc.UpdateProfile(r.Context(), cookie.Value, clientIP(r), in)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest) // Simple mapping for now
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProfileHandler) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		jsonError(w, "non authentifié", http.StatusUnauthorized)
		return
	}

	// Max 2MB
	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
	file, header, err := r.FormFile("avatar")
	if err != nil {
		jsonError(w, "invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	path, err := h.uploadSvc.UploadAvatar(r.Context(), cookie.Value, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"avatar_path": path})
}
