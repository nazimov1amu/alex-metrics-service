package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type AgentConfig struct {
	Address string `env:"ADDRESS"`
	PollInterval time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}


func NewAgentConfig() *AgentConfig {
	var (
		address string
		pollInterval time.Duration
		reportInterval time.Duration
	)

	config := &AgentConfig{}

	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	flag.StringVar(&address, "a", "localhost:8080", "HTTP server address host:port")
	flag.DurationVar(&pollInterval, "p", 2*time.Second, "poll interval")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "report interval")
	flag.Parse()

	if err := env.Parse(config); err != nil {
		log.Fatalf("failed to parse env config: %v", err)
	}

	if config.Address == "" {
		config.Address = address
	}
	if config.PollInterval == 0 {
		config.PollInterval = pollInterval
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = reportInterval
	}

	return config
}