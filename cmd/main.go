package main

import (
	"ServiceManager/internal/app"
	"ServiceManager/internal/config"
	"ServiceManager/pkg/utils"
	"context"
	"log"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: utils.GetSlogLevelByName(cfg.LogLevel),
	}))

	ctx = context.WithValue(ctx, "logger", logger)
	ctx = context.WithValue(ctx, "config", cfg)

	application := app.NewApp(ctx)

	if err = application.Start(ctx); err != nil {
		logger.Error("application error", "error", err.Error())
	}
	logger.Info("application shutdown successfully")
}
