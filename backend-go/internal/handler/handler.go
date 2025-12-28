package handler

import (
	"net/http"
	"time"

	"plat-detection-system/backend-go/internal/config"
	"plat-detection-system/backend-go/internal/httpapi"
	"plat-detection-system/backend-go/internal/service"
)

type Handler struct {
	cfg       config.Config
	detectSvc *service.DetectService
}

func New(cfg config.Config, detectSvc *service.DetectService) *Handler {
	return &Handler{cfg: cfg, detectSvc: detectSvc}
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"request_id": httpapi.RequestID(r.Context()),
	})
}

func (h *Handler) Detect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	imagePath, cleanup, err := httpapi.ParseImageUpload(w, r, h.cfg)
	if err != nil {
		httpapi.WriteError(w, r.Context(), start, err)
		return
	}
	if cleanup != nil {
		defer cleanup()
	}

	if h.detectSvc == nil {
		httpapi.WriteError(w, r.Context(), start, nil)
		return
	}

	data, meta, err := h.detectSvc.Detect(r.Context(), imagePath)
	if err != nil {
		httpapi.WriteError(w, r.Context(), start, err)
		return
	}

	httpapi.WriteSuccess(w, r.Context(), start, data, meta)
}
