package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validToken           = "valid-token-123"
	unauthorizedResponse = "unauthorized\n"
	okResponse           = "OK"
)

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		expectedToken  string
		headerToken    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "should allow request with valid token",
			expectedToken:  validToken,
			headerToken:    "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedBody:   okResponse,
		},
		{
			name:           "should reject request without authorization header",
			expectedToken:  validToken,
			headerToken:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   unauthorizedResponse,
		},
		{
			name:           "should reject request with invalid token",
			expectedToken:  validToken,
			headerToken:    "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   unauthorizedResponse,
		},
		{
			name:           "should reject request with malformed authorization header",
			expectedToken:  validToken,
			headerToken:    "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   unauthorizedResponse,
		},
		{
			name:           "should reject when token is missing from header",
			expectedToken:  validToken,
			headerToken:    "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   unauthorizedResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(okResponse))
			})

			// Wrap with auth middleware
			middleware := AuthMiddleware(tt.expectedToken, logger)
			wrappedHandler := middleware(testHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.headerToken != "" {
				req.Header.Set("Authorization", tt.headerToken)
			}
			w := httptest.NewRecorder()

			// Act
			wrappedHandler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}
