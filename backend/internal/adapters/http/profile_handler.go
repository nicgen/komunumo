package http

import (
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
}

func (h *ProfileHandler) HandleUpdateMyProfile(w http.ResponseWriter, r *http.Request) {
}

func (h *ProfileHandler) HandleUploadAvatar(w http.ResponseWriter, r *http.Request) {
}
