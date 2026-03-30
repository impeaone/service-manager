package middleware

import (
	"ServiceManager/internal/domain"
	service "ServiceManager/internal/service/auth"
	"ServiceManager/pkg/utils"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Middleware func(http.Handler) http.Handler

type contextKey string

const UserContextKey contextKey = "user"

func Logger(logs *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.String(), "static") && !strings.Contains(r.URL.String(), "swagger") {
				logs.Info("request url: "+r.URL.String(), "client", r.RemoteAddr, "method", r.Method,
					"time", time.Now().String(), "place", utils.GetPlace())
			}
			next.ServeHTTP(w, r)
		})
	}
}

func PanicMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic middleware", "panic", err.(error).Error(), utils.GetPlace())

					utils.SendJSON(w, map[string]interface{}{
						"error":   err.(error).Error(),
						"message": "internal server error",
					}, http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func AuthMiddleware(authService *service.AuthService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			user, err := authService.VerifyToken(r.Context(), token)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) (*domain.User, error) {
	user, ok := ctx.Value(UserContextKey).(*domain.User)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}
