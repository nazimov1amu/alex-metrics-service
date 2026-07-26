package service

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

type AgentService struct {
	config *config.Config
}

func NewAgentService(config *config.Config) *AgentService {
	return &AgentService{config: config}
}

func (s *AgentService) collectRuntimeMetrics() map[string]float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m) 
	return map[string]float64{
		"Alloc":         float64(m.Alloc),
		"BuckHashSys":   float64(m.BuckHashSys),
		"Frees":         float64(m.Frees),
		"GCCPUFraction": m.GCCPUFraction,
		"GCSys":         float64(m.GCSys),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapObjects":   float64(m.HeapObjects),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapSys":       float64(m.HeapSys),
		"LastGC":        float64(m.LastGC),
		"Lookups":       float64(m.Lookups),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"Mallocs":       float64(m.Mallocs),
		"NextGC":        float64(m.NextGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"NumGC":         float64(m.NumGC),
		"OtherSys":      float64(m.OtherSys),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"Sys":           float64(m.Sys),
		"TotalAlloc":    float64(m.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}
}


func (s *AgentService) updatePollCount() error {
	endpoint := fmt.Sprintf("http://%s/update/%s/%s/%s", s.config.Address, model.Counter, "PollCount", "1")
	resp, err := http.Post(endpoint, "application/plain", bytes.NewBufferString("1"))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update poll count: %s", resp.Status)
	}
	return nil
}


func (s *AgentService) sendMetrics(metrics map[string]float64) error {
	for name, value := range metrics {
		endpoint := fmt.Sprintf("http://%s/update/%s/%s/%s", s.config.Address, model.Gauge, name, strconv.FormatFloat(value, 'f', -1, 64))
	
		input := model.MetricsInput{
			Name: name,
			RawValue: strconv.FormatFloat(value, 'f', -1, 64),
			MType: model.Gauge,
		}
		err := input.Validate()
		if err != nil {
			return err
		}

		resp, err := http.Post(endpoint, "application/plain", bytes.NewBufferString(input.RawValue))
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to send metrics: %s", resp.Status)
		}

		err = s.updatePollCount()
		if err != nil {
			return err
		}

		time.Sleep(s.config.ReportInterval)
	}
	return nil
}

func (s *AgentService) Run() {
	fmt.Println("Starting agent service")
	for {
		metrics := s.collectRuntimeMetrics()
		if err := s.sendMetrics(metrics); err != nil {
			fmt.Printf("failed to send metrics: %v\n", err)
		}
		time.Sleep(s.config.PollInterval)
	}
}