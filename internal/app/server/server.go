package server

import (
	"ServiceManager/internal/config"
	"ServiceManager/internal/middleware"
	service "ServiceManager/internal/service/auth"
	"ServiceManager/internal/transport/handler"
	"context"
	"log/slog"
	"net/http"
)

type WebHookServer struct {
	handlers *handler.APIHandler
	server   *http.Server
	logger   *slog.Logger
	cfg      *config.Config
	ctx      context.Context
}

func NewWebHookServer(ctx context.Context, authService *service.AuthService, authHandls *handler.AuthHandler, handls *handler.APIHandler) *WebHookServer {
	logger := ctx.Value("logger").(*slog.Logger)
	cfg := ctx.Value("config").(*config.Config)

	auth := http.NewServeMux()
	auth.HandleFunc("POST /api/auth/register/request", authHandls.RequestRegistration)
	auth.HandleFunc("POST /api/auth/register/verify", authHandls.VerifyRegistration)
	auth.HandleFunc("POST /api/auth/login/request", authHandls.RequestLogin)
	auth.HandleFunc("POST /api/auth/login/verify", authHandls.VerifyLogin)
	auth.HandleFunc("POST /api/auth/refresh", authHandls.Refresh)
	auth.HandleFunc("POST /api/auth/verify", authHandls.Verify)

	router := http.NewServeMux()

	router.HandleFunc("POST /api/auth/logout", authHandls.Logout)

	router.HandleFunc("POST /api/service", handls.AddService)
	router.HandleFunc("GET /api/service/{service_id}", handls.GetService)
	router.HandleFunc("PUT /api/service", handls.UpdateService)
	router.HandleFunc("DELETE /api/service/{service_id}", handls.DeleteService)
	router.HandleFunc("GET /api/services", handls.GetServices)
	router.HandleFunc("POST /api/services/execute", handls.ExecuteWebHook)

	auth.Handle("/api/", middleware.AuthMiddleware(authService)(router))

	routerWithMiddleware := middleware.PanicMiddleware(logger)(
		middleware.Logger(logger)(auth),
	)

	server := &http.Server{
		Addr:    cfg.ServerEndpoint,
		Handler: routerWithMiddleware,
	}

	return &WebHookServer{
		handlers: handls,
		server:   server,
		logger:   logger,
		ctx:      ctx,
		cfg:      cfg,
	}
}

func (s *WebHookServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *WebHookServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
