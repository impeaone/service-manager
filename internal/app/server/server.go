package server

import (
	"ServiceManager/internal/config"
	"ServiceManager/internal/middleware"
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

func NewWebHookServer(ctx context.Context, handls *handler.APIHandler) *WebHookServer {
	logger := ctx.Value("logger").(*slog.Logger)
	cfg := ctx.Value("config").(*config.Config)

	router := http.NewServeMux()

	router.HandleFunc("POST /api/service", handls.AddService)
	router.HandleFunc("GET /api/service/{service_id}", handls.GetService)
	router.HandleFunc("PUT /api/service", handls.UpdateService)
	router.HandleFunc("DELETE /api/service/{service_id}", handls.DeleteService)
	router.HandleFunc("GET /api/services", handls.GetServices)

	router.HandleFunc("POST /api/services/execute", handls.ExecuteWebHook)

	router.HandleFunc("POST /api/auth/register", func(writer http.ResponseWriter, request *http.Request) {})
	router.HandleFunc("POST /api/auth/login", func(writer http.ResponseWriter, request *http.Request) {})
	router.HandleFunc("POST /api/auth/refresh", func(writer http.ResponseWriter, request *http.Request) {})
	router.HandleFunc("POST /api/auth/logout", func(writer http.ResponseWriter, request *http.Request) {})
	router.HandleFunc("POST /api/auth/verify", func(writer http.ResponseWriter, request *http.Request) {})

	routerWithMiddleware := middleware.PanicMiddleware(logger)(
		middleware.Logger(logger)(router),
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
