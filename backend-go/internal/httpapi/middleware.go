package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += int64(n)
	return n, err
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := RequestIDFromHeader(r)
		if reqID == "" {
			reqID = NewRequestID()
		}
		w.Header().Set("X-Request-Id", reqID)

		l := slog.Default().With(
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		ctx := WithRequestID(r.Context(), reqID)
		ctx = WithLogger(ctx, l)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		Logger(r.Context()).Info("request start")
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		Logger(r.Context()).Info(
			"request end",
			slog.Int("status", sw.status),
			slog.Int64("bytes", sw.bytes),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				Logger(r.Context()).Error("panic recovered", slog.Any("panic", rec))
				WriteError(w, r.Context(), time.Now(), nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
