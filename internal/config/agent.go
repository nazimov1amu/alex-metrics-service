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
		pollInterval int
		reportInterval int
	)

	config := &AgentConfig{}

	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	flag.StringVar(&address, "a", "localhost:8080", "HTTP server address host:port")
	flag.IntVar(&pollInterval, "p", 2, "poll interval")
	flag.IntVar(&reportInterval, "r", 10, "report interval")
	flag.Parse()

	if err := env.Parse(config); err != nil {
		log.Fatalf("failed to parse env config: %v", err)
	}

	if config.Address == "" {
		config.Address = address
	}
	if config.PollInterval == 0 {
		config.PollInterval = time.Duration(pollInterval) * time.Second
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = time.Duration(reportInterval) * time.Second	
	}

	return config
}