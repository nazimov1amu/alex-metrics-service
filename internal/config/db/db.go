package db

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	DatabaseDSN string `env:"DATABASE_DSN"`
}

func NewDatabaseConfig() *DatabaseConfig {
	cfg := &DatabaseConfig{}

	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database DSN")

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	if v, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DatabaseDSN = v
	}

	return cfg
}