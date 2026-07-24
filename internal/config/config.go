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
	var cfg Config
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address to listen")
	flag.DurationVar(&cfg.PollInterval, "p", 2*time.Second, "poll interval (e.g. 2s)")
	flag.DurationVar(&cfg.ReportInterval, "r", 10*time.Second, "report interval (e.g. 10s)")
	flag.Parse()
	return &cfg
}