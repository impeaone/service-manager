package main

import (
	"ServiceManager/internal/app"
	"ServiceManager/internal/config"
	"ServiceManager/pkg/logger"
	"context"
	"log"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logFile, logs := logger.NewSlogLogger(cfg)
	defer func() { _ = logFile.Close() }()

	ctx = context.WithValue(ctx, "logger", logs)
	ctx = context.WithValue(ctx, "config", cfg)

	application := app.NewApp(ctx)

	if err = application.Start(ctx); err != nil {
		logs.Error("application error", "error", err.Error())
	}
	logs.Info("application shutdown successfully")
}
