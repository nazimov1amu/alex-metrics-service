// Package agent запускает сбор runtime-метрик и их отправку на сервер.
package agent

import (
	"os"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
)

// Agent собирает метрики процесса и отправляет их HTTP-серверу.
type Agent struct {
	agentService *service.AgentService
}

// NewAgent создаёт агент с заданным адресом сервера и интервалами опроса.
func NewAgent(cfg *config.AgentConfig) *Agent {
	return &Agent{agentService: service.NewAgentService(cfg)}
}

// Run запускает сбор и отправку метрик и блокируется до сигнала в quit.
func (a *Agent) Run(quit chan os.Signal) {
	a.agentService.Run(quit)
}
