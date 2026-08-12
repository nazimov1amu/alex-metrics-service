// internal/config/config.go
package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Address        string `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore bool `env:"RESTORE"`
}


func NewServerConfig() *ServerConfig {
	var (
		address        string    
		storeInterval  int
		fileStoragePath string
		restore bool
	)

	config := &ServerConfig{}

	flag.StringVar(&address, "a", "localhost:8080", "HTTP server address host:port")
	flag.IntVar(&storeInterval, "i", 300, "store interval")
	flag.StringVar(&fileStoragePath, "f", "storage.json", "file storage path")
	flag.BoolVar(&restore, "r", false, "restore from file")
	flag.Parse()

	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	if err := env.Parse(config); err != nil {
		log.Fatalf("failed to parse env config: %v", err)
	}

	if config.Address == "" {
		config.Address = address
	}
	if config.StoreInterval == 0 {
		config.StoreInterval = time.Duration(storeInterval) * time.Second
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = fileStoragePath
	}
	if config.Restore == false {
		config.Restore = restore
	}

	return config
}
