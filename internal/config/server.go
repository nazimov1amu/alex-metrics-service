package config

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Address         string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

func NewServerConfig() *ServerConfig {
	config := &ServerConfig{}

	flag.StringVar(&config.Address, "a", "localhost:8080", "HTTP server address host:port")
	flag.IntVar(&config.StoreInterval, "i", 300, "store interval")
	flag.StringVar(&config.FileStoragePath, "f", "storage.json", "file storage path")
	flag.BoolVar(&config.Restore, "r", false, "restore from file")
	flag.Parse()

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}

	if address, ok := os.LookupEnv("ADDRESS"); ok {
		config.Address = address
	}

	if storeIntervalEnv, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		storeInterval, err := strconv.Atoi(storeIntervalEnv)
		if err != nil {
			log.Fatalf("failed to convert STORE_INTERVAL to int: %v", err)
		}
		config.StoreInterval = storeInterval
	}

	if fileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		config.FileStoragePath = fileStoragePath
	}

	if restoreEnv, ok := os.LookupEnv("RESTORE"); ok {
		restore, err := strconv.ParseBool(restoreEnv)
		if err != nil {
			log.Fatalf("failed to convert RESTORE to bool: %v", err)
		}
		config.Restore = restore
	}

	return config
}
