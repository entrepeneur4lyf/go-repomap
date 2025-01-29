package middleware

import (
	"net/http"
	"strings"

	"example.com/web/services"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware() func(http.Handler) http.Handler {
	middleware := &AuthMiddleware{
		authService: services.NewAuthService(),
	}
	return middleware.Handle
}

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		if !m.authService.ValidateToken(token) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
