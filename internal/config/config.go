// internal/config/config.go
package config

import (
	"flag"
	"time"
)

type Config struct {
	Address        string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func NewConfig() *Config {
	var (
		address        string
		pollInterval   int
		reportInterval int
	)

	flag.StringVar(&address, "a", "localhost:8080", "HTTP server address host:port")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.Parse()

	return &Config{
		Address:        address,
		PollInterval:   time.Duration(pollInterval) * time.Second,
		ReportInterval: time.Duration(reportInterval) * time.Second,
	}
}
