package config

import (
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	DatabaseDSN string
	ServerAddr  string
}

func Load(logger *zap.Logger) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		logger.Warn("Failed to load .env file, using default env variable")
	}	

	cfg := &Config{
		DatabaseDSN: os.Getenv("DB_DSN"),
		ServerAddr:  os.Getenv("SERVER_ADDR"),
	}
	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN="user:password@tcp(localhost:3306)/order_matching?parseTime=true"
	}
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = ":8080"
	}
	return cfg, nil
}

