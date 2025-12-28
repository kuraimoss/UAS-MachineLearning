package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"plat-detection-system/backend-go/internal/errx"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope[T any] struct {
	Status           string    `json:"status"`
	RequestID        string    `json:"request_id"`
	ProcessingTimeMs int64     `json:"processing_time_ms"`
	Data             *T        `json:"data,omitempty"`
	Error            *ErrorBody `json:"error,omitempty"`
	Meta             any       `json:"meta,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteSuccess[T any](w http.ResponseWriter, ctx context.Context, start time.Time, data T, meta any) {
	resp := Envelope[T]{
		Status:           "success",
		RequestID:        RequestID(ctx),
		ProcessingTimeMs: time.Since(start).Milliseconds(),
		Data:             &data,
		Meta:             meta,
	}
	WriteJSON(w, http.StatusOK, resp)
}

func WriteError(w http.ResponseWriter, ctx context.Context, start time.Time, err error) {
	app, ok := errx.As(err)
	if !ok || app == nil {
		app = errx.New("INTERNAL_ERROR", "internal error", http.StatusInternalServerError)
	}

	resp := Envelope[any]{
		Status:           "error",
		RequestID:        RequestID(ctx),
		ProcessingTimeMs: time.Since(start).Milliseconds(),
		Error: &ErrorBody{
			Code:    app.Code,
			Message: app.Message,
		},
	}
	WriteJSON(w, app.HTTPStatus, resp)
}
