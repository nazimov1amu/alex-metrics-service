package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
)

type AgentService struct {
	config    *config.Config
	metrics   map[string]float64
	mu        sync.Mutex
	pollCount int64
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

func (s *AgentService) postMetric(metric model.Metrics) error {
	body, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s/update/", s.config.Address)
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send metric %s: %s", metric.ID, resp.Status)
	}
	return nil
}

func (s *AgentService) sendMetrics(metrics map[string]float64, pollCount int64) error {
	for name, value := range metrics {
		v := value
		if err := s.postMetric(model.Metrics{
			ID:    name,
			MType: model.Gauge,
			Value: &v,
		}); err != nil {
			log.Printf("failed to send metrics %s: %v\n", name, err)
			return err
		}
	}

	delta := pollCount
	if err := s.postMetric(model.Metrics{
		ID:    "PollCount",
		MType: model.Counter,
		Delta: &delta,
	}); err != nil {
		log.Printf("failed to update poll count: %v\n", err)
		return err
	}
	return nil
}

func (s *AgentService) pollLoop() {
	for {
		time.Sleep(s.config.PollInterval)
		metrics := s.collectRuntimeMetrics()
		s.mu.Lock()
		s.pollCount++
		s.metrics = metrics
		s.mu.Unlock()
	}
}

func (s *AgentService) reportLoop() {
	for {
		time.Sleep(s.config.ReportInterval)
		s.mu.Lock()
		snapshot := maps.Clone(s.metrics)
		pollCount := s.pollCount
		s.pollCount = 0
		s.mu.Unlock()
		if err := s.sendMetrics(snapshot, pollCount); err != nil {
			log.Printf("failed to send metrics: %v\n", err)
		}
	}
}

func (s *AgentService) Run() {
	s.metrics = make(map[string]float64)
	go s.pollLoop()
	go s.reportLoop()
	select {}
}
