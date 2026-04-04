package logger

import (
	"ServiceManager/internal/config"
	"ServiceManager/pkg/utils"
	"io"
	"log/slog"
	"os"
)

const logFile = "./app.log"

func NewSlogLogger(cfg *config.Config) (*os.File, *slog.Logger) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)

	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: utils.GetSlogLevelByName(cfg.LogLevel),
	}))
	return file, logger
}
