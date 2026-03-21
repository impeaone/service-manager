package utils

import (
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func SendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func ReadCloserToJSONMap(reader io.ReadCloser) (map[string]interface{}, error) {
	var bodyReq map[string]interface{}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &bodyReq); err != nil {
		return bodyReq, err
	}
	return bodyReq, nil
}

func GetPlace() string {
	_, file, line, _ := runtime.Caller(1)
	split := strings.Split(file, "/")
	StartFile := split[len(split)-1]
	place := StartFile + ":" + strconv.Itoa(line)
	return place
}

func GetSlogLevelByName(logLevel string) slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}

}
