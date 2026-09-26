package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Alexunder2003/alex-metrics-service/internal/agent"
	"github.com/Alexunder2003/alex-metrics-service/internal/config"
)

func main() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	cfg := config.NewAgentConfig()
	agent.NewAgent(cfg).Run(quit)
}
