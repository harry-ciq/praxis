package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/praxis-social/praxis/server/internal/service"
)

type contextKey string

const userClaimsKey contextKey = "userClaims"

// AuthMiddleware validates the Bearer token from the Authorization header
// and stores user claims in the request context.
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := authService.ValidateAccessToken(parts[1])
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TryAuthMiddleware attempts to authenticate the request but does not reject
// unauthenticated requests. If a valid token is present, claims are stored in
// the context; otherwise the request proceeds with no claims.
func TryAuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
					claims, err := authService.ValidateAccessToken(parts[1])
					if err == nil {
						ctx := context.WithValue(r.Context(), userClaimsKey, claims)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext retrieves the user claims from the request context.
func GetUserFromContext(ctx context.Context) *service.Claims {
	claims, ok := ctx.Value(userClaimsKey).(*service.Claims)
	if !ok {
		return nil
	}
	return claims
}
