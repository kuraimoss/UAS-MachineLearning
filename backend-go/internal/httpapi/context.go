package httpapi

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
)

type ctxKey string

const (
	ctxKeyRequestID ctxKey = "request_id"
	ctxKeyLogger    ctxKey = "logger"
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, id)
}

func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyRequestID).(string); ok {
		return v
	}
	return ""
}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, l)
}

func Logger(ctx context.Context) *slog.Logger {
	if v, ok := ctx.Value(ctxKeyLogger).(*slog.Logger); ok && v != nil {
		return v
	}
	return slog.Default()
}

func NewRequestID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return "req-unknown"
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}

func RequestIDFromHeader(r *http.Request) string {
	id := r.Header.Get("X-Request-Id")
	if id == "" {
		id = r.Header.Get("X-Request-ID")
	}
	return id
}

