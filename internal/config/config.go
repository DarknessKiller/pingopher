package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env        string
	Host       string
	Port       string
	SQLitePath string
}

func Load() (*Config, error) {
	env := os.Getenv("ENV")

	if env == "" {
		if err := godotenv.Load(".env"); err != nil {
			panic("Could not load .env file")
		}
	}

	cfg := &Config{
		Env:        os.Getenv("ENV"),
		Host:       os.Getenv("PINGOPHER_HOST"),
		Port:       os.Getenv("PINGOPHER_PORT"),
		SQLitePath: os.Getenv("PINGOPHER_DB_PATH"),
	}

	return cfg, nil
}
