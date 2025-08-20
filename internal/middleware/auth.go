package middleware

import (
	"log/slog"
	"net/http"
	"strings"
)

// AuthMiddleware valida Bearer token.
func AuthMiddleware(expectedToken string, logger *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			parts := strings.Split(auth, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || parts[1] != expectedToken {
				logger.Warn("Acesso não autorizado", "path", r.URL.Path)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
}
