package agent

import (
	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
)

type Agent struct {
	agentService *service.AgentService
}

func NewAgent(cfg *config.Config) *Agent {
	return &Agent{agentService: service.NewAgentService(config.NewConfig())}
}

func (a *Agent) Run() {
	a.agentService.Run()
}