package main

import (
	"ServiceManager/internal/app"
	"ServiceManager/internal/config"
	"ServiceManager/pkg/utils"
	"context"
	"io"
	"log"
	"log/slog"
	"os"
)

const logFile = "app.log"

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer func() { _ = file.Close() }()

	multiWriter := io.MultiWriter(file, os.Stdout)

	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
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
