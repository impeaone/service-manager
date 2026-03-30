package app

import (
	"ServiceManager/internal/app/server"
	"ServiceManager/internal/config"
	"ServiceManager/internal/repository/postgres"
	redisRepo "ServiceManager/internal/repository/redis"
	service "ServiceManager/internal/service/auth"
	"ServiceManager/internal/service/email"
	"ServiceManager/internal/service/jwt"
	"ServiceManager/internal/service/services"
	"ServiceManager/internal/transport/handler"
	"ServiceManager/migration"
	"ServiceManager/pkg/closer"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const ShutDownTimeOut = time.Second * 30

type App struct {
	serv   *server.WebHookServer
	logger *slog.Logger
	closer *closer.Closer
	cfg    *config.Config
}

func NewApp(ctx context.Context) *App {
	logger := ctx.Value("logger").(*slog.Logger)
	cfg := ctx.Value("config").(*config.Config)

	clsr := closer.NewCloser(logger)

	if err := migration.Migrate(ctx); err != nil && !errors.As(err, &migration.ErrorNoChange) {
		logger.Error("migration migrate failed", "err", err)
		return nil
	}
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Репозитории Postgres/Redis   ////////////////////////////////////////////////////////////////////////////////////
	pool, err := postgres.InitPGPool(ctx)
	if err != nil {
		logger.Error("init postgres pool failed", "err", err)
		return nil
	}
	clsr.AddFunc("repository", pool.Close)

	serviceRepo := postgres.NewServiceRepository(pool)
	userRepo := postgres.NewUserRepository(pool)

	//TODO: config доделать
	redisClient, err := redisRepo.NewClient("192.168.3.92:6379", "", 0)
	if err != nil {
		logger.Error("init redis client failed", "err", err)
		return nil
	}
	clsr.Add("redis repo", redisClient.Close)

	otpRepo := redisRepo.NewOTPRepository(redisClient.Client)
	tokenRepo := redisRepo.NewTokenRepository(redisClient.Client)

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Сервисы   ///////////////////////////////////////////////////////////////////////////////////////////////////////
	//TODO: конфиг
	jwtService := jwt.NewJWTService("aedfqefqw", 1*time.Hour, 2*time.Hour)
	emailService := email.NewMockEmailService()

	servic := services.NewServiceManager(serviceRepo)
	authService := service.NewAuthService(userRepo, otpRepo, tokenRepo, jwtService, emailService)

	serviceHandlers := handler.NewAPIHandler(ctx, servic)
	authHandlers := handler.NewAuthHandler(authService)

	serv := server.NewWebHookServer(ctx, authService, authHandlers, serviceHandlers)
	clsr.Add("server", serv.Shutdown)

	return &App{
		serv:   serv,
		logger: logger,
		closer: clsr,
		cfg:    cfg,
	}
}

func (app *App) Start(ctx context.Context) error {
	if app == nil {
		return fmt.Errorf("app is nil")
	}

	errCh := make(chan error, 2)

	app.logger.Info("starting server", "addr", app.cfg.ServerEndpoint)
	go func() {
		errCh <- app.serv.Start()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigChan:
		app.logger.Info("signal received", "signal", sig.String())

		ctxx, cancel := context.WithTimeout(ctx, ShutDownTimeOut)
		defer cancel()

		endChan := make(chan error, 1)

		go func() {
			endChan <- app.Shutdown(ctxx)
		}()

		select {
		case err := <-endChan:
			return err
		case <-ctxx.Done():
			return ctxx.Err()
		}
	}
}

func (app *App) Shutdown(ctx context.Context) error {
	return app.closer.Close(ctx)
}
