package server

import (
	"ServiceManager/internal/config"
	"ServiceManager/internal/middleware"
	"ServiceManager/internal/transport/handler"
	"context"
	"log/slog"
	"net/http"
	"sync"
)

type WebHookServer struct {
	handlers *handler.APIHandler
	server   *http.Server
	logger   *slog.Logger
	cfg      *config.Config
	closeCh  chan struct{}
	cons     *sync.WaitGroup
	ctx      context.Context
}

func NewWebHookServer(ctx context.Context, handls *handler.APIHandler) *WebHookServer {
	logger := ctx.Value("logger").(*slog.Logger)
	cfg := ctx.Value("config").(*config.Config)

	closeChan := make(chan struct{})
	cons := new(sync.WaitGroup)

	router := http.NewServeMux()
	router.HandleFunc("GET /api/services", handls.GetServices)
	router.HandleFunc("GET /api/service/{service_id}", handls.GetService)
	router.HandleFunc("DELETE /api/service/{service_id}", handls.DeleteService)
	router.HandleFunc("POST /api/services", handls.AddService)
	router.HandleFunc("GET /api/services/execute", handls.ExecuteWebHook)

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
		closeCh:  closeChan,
		cons:     cons,
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
