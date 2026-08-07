// internal/config/config.go
package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}

type ConfigVars struct {
	Address        string
	PollInterval   int
	ReportInterval int
}

func NewConfig() *Config {
	var (
		address        string    
		pollInterval   int    
		reportInterval int
	)

	configVars := &ConfigVars{}

	flag.StringVar(&address, "a", "localhost:8080", "HTTP server address host:port")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.Parse()

	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	if err := env.Parse(configVars); err != nil {
		log.Fatalf("failed to parse env config: %v", err)
	}

	if configVars.Address == "" {
		configVars.Address = address
	}
	if configVars.PollInterval == 0 {
		configVars.PollInterval = pollInterval
	}
	if configVars.ReportInterval == 0 {
		configVars.ReportInterval = reportInterval
	}

	return &Config{
		Address:        configVars.Address,
		PollInterval:   time.Duration(configVars.PollInterval) * time.Second,
		ReportInterval: time.Duration(configVars.ReportInterval) * time.Second,
	}
}
