package agent

import (
	"os"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
)

type Agent struct {
	agentService *service.AgentService
}

func NewAgent(cfg *config.AgentConfig) *Agent {
	return &Agent{agentService: service.NewAgentService(cfg)}
}

func (a *Agent) Run(quit chan os.Signal) {
	a.agentService.Run(quit)
}
