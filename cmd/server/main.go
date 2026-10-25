package main

import (
	"log"

	"github.com/Alexunder2003/alex-metrics-service/internal/app"
)

func main() {
	application := app.NewApp()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}

	if err := application.Close(); err != nil {
		log.Printf("failed to close application: %v", err)
	}
}
