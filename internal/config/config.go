package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type CloudflareD1Config struct {
	AccountID      string
	AuthToken      string
	DatabaseString string
}

type Config struct {
	Env          string
	Host         string
	Port         string
	DatabaseType string

	RedisHost     string
	RedisPort     string
	RedisPassword string

	// Databases
	SQLitePath   string
	CloudflareD1 CloudflareD1Config
}

func Load() (*Config, error) {
	env := os.Getenv("ENV")

	if env == "" {
		if err := godotenv.Load(".env"); err != nil {
			panic("Could not load .env file")
		}
	}

	cfg := &Config{
		Env:           os.Getenv("ENV"),
		Host:          os.Getenv("PINGOPHER_HOST"),
		Port:          os.Getenv("PINGOPHER_PORT"),
		DatabaseType:  strings.ToLower(os.Getenv("PINGOPHER_DB_TYPE")),
		SQLitePath:    os.Getenv("PINGOPHER_DB_PATH"),
		RedisHost:     os.Getenv("PINGOPHER_REDIS_HOST"),
		RedisPort:     os.Getenv("PINGOPHER_REDIS_PORT"),
		RedisPassword: os.Getenv("PINGOPHER_REDIS_PASSWORD"),
		CloudflareD1: CloudflareD1Config{
			AccountID:      os.Getenv("PINGOPHER_CF_D1_ACCOUNT_ID"),
			AuthToken:      os.Getenv("PINGOPHER_CF_D1_AUTH_TOKEN"),
			DatabaseString: os.Getenv("PINGOPHER_CF_D1_DATABASE_STRING"),
		},
	}

	return cfg, nil
}
