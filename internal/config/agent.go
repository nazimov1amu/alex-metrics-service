package config

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type AgentConfig struct {
	Address        string
	PollInterval   int
	ReportInterval int
}

func NewAgentConfig() *AgentConfig {
	config := &AgentConfig{}

	flag.StringVar(&config.Address, "a", "localhost:8080", "HTTP server address host:port")
	flag.IntVar(&config.PollInterval, "p", 2, "poll interval")
	flag.IntVar(&config.ReportInterval, "r", 10, "report interval")
	flag.Parse()

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	if address, ok := os.LookupEnv("ADDRESS"); ok {
		config.Address = address
	}

	if pollIntervalEnv, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		pollInterval, err := strconv.Atoi(pollIntervalEnv)
		if err != nil {
			log.Fatalf("failed to convert POLL_INTERVAL to int: %v", err)
		}
		config.PollInterval = pollInterval
	}

	if reportIntervalEnv, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		reportInterval, err := strconv.Atoi(reportIntervalEnv)
		if err != nil {
			log.Fatalf("failed to convert REPORT_INTERVAL to int: %v", err)
		}
		config.ReportInterval = reportInterval
	}

	return config
}
