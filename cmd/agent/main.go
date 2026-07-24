package main

import (
	"github.com/Alexunder2003/alex-metrics-service/internal/agent"
	"github.com/Alexunder2003/alex-metrics-service/internal/config"
)

func main() {
	cfg := config.NewConfig()
	agent.NewAgent(cfg).Run()
}
