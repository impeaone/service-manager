package config

import (
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"log"
)

const envFile = "./internal/config/.env"

type Config struct {
	// Server
	ServerEndpoint string `env:"SERVER_ENDPOINT" envDefault:":8080"`

	DBType string `env:"DB_TYPE" envDefault:"sqlite3"`
	// DBHost - postgres host
	DBHost string `env:"DB_HOST" envDefault:"localhost"`
	// DBHost - postgres port
	DBPort int `env:"DB_PORT" envDefault:"5432"`
	// DBHost - postgres user
	DBUser string `env:"DB_USER" envDefault:"postgres"`
	// DBHost - postgres password
	DBPassword string `env:"DB_PASSWORD" envDefault:"password"`
	// DBHost - postgres name
	DBName string `env:"DB_NAME" envDefault:"myapp"`
	// DBHost - postgres ssl mode
	DBSSLMode string `env:"DB_SSL_MODE" envDefault:"disable"`

	// DBFilePath - path to SQLITE db file
	DBFilePath string `env:"DB_FILE_PATH" envDefault:"./database.db"`

	// Logger
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

// LoadConfig загружает конфигурацию из .env файла и переменных окружения
// Cначала подгружается .env (в переменные окружения заносятся данные) файл если он есть.
// Потом уже из переменных окружения достается нужное, если нет чего-то default используется
func LoadConfig() (*Config, error) {

	if err := godotenv.Load(envFile); err != nil {
		log.Println("No .env file found or error loading it")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
