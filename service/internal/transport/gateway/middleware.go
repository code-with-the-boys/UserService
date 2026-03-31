package gateway

import (
	"net/http"
	"strings"

	customContext "github.com/code-with-the-boys/UserService/internal/context"
	"github.com/code-with-the-boys/UserService/pkg/auth"
	"go.uber.org/zap"
)

// JWTHTTPMiddleware validates Bearer JWT for non-public routes and sets auth on request context.
func JWTHTTPMiddleware(logger *zap.Logger, jwt auth.JwtAuth, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		path := r.URL.Path
		if IsPublicHTTPRoute(r.Method, path) {
			next.ServeHTTP(w, r)
			return
		}

		authHeaders, ok := r.Header["Authorization"]
		if !ok || len(authHeaders) == 0 {
			http.Error(w, `{"error":"missing authorization"}`, http.StatusUnauthorized)
			return
		}
		tokenRaw := strings.TrimSpace(authHeaders[0])
		const prefix = "Bearer "
		if len(tokenRaw) < len(prefix) || !strings.EqualFold(tokenRaw[:len(prefix)], prefix) {
			http.Error(w, `{"error":"invalid authorization header"}`, http.StatusUnauthorized)
			return
		}
		token := strings.TrimSpace(tokenRaw[len(prefix):])
		if token == "" {
			http.Error(w, `{"error":"empty token"}`, http.StatusUnauthorized)
			return
		}
		user, err := jwt.ValidateToken(token)
		if err != nil {
			if logger != nil {
				logger.Debug("http jwt validation failed", zap.Error(err))
			}
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}
		ctx := customContext.WithAuth(r.Context(), customContext.AuthData{
			User:     user,
			RawToken: token,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsPublicHTTPRoute matches grpc-gateway paths that skip JWT on HTTP edge.
func IsPublicHTTPRoute(method, path string) bool {
	if method == http.MethodGet && path == "/api/v1/train/health" {
		return true
	}
	switch path {
	case "/api/v1/auth/login", "/api/v1/auth/refresh":
		return method == http.MethodPost
	case "/api/v1/users":
		return method == http.MethodPost // SignUp
	default:
		return false
	}
}
