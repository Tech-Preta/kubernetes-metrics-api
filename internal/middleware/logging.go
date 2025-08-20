package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingMiddleware loga requisições.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &respWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)
			logger.Info("http", "method", r.Method, "path", r.URL.Path, "status", rw.status, "dur", time.Since(start))
		})
	}
}

type respWriter struct {
	http.ResponseWriter
	status int
}

func (r *respWriter) WriteHeader(code int) { r.status = code; r.ResponseWriter.WriteHeader(code) }
