package middleware

import (
	"ServiceManager/pkg/utils"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Middleware func(http.Handler) http.Handler

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
