package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/encoding"
	"github.com/Alexunder2003/alex-metrics-service/internal/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type AgentService struct {
	config    *config.AgentConfig
	metrics   map[string]float64
	mu        sync.Mutex
	pollCount int64
}

func NewAgentService(config *config.AgentConfig) *AgentService {
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
		"RandomValue":   rand.Float64(),
	}
}

func (s *AgentService) collectAdditionalMetrics() (map[string]float64, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	percents, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, err
	}

	out := map[string]float64{
		"TotalMemory": float64(vm.Total),
		"FreeMemory":  float64(vm.Free),
	}
	for i, p := range percents {
		out[fmt.Sprintf("CPUutilization%d", i+1)] = p
	}
	return out, nil
}

func (s *AgentService) additionalPollLoop() {
	ticker := time.NewTicker(time.Duration(s.config.PollInterval) * time.Second)
	for range ticker.C {
		extra, err := s.collectAdditionalMetrics()
		if err != nil {
			log.Printf("failed to collect additional metrics: %v", err)
			continue
		}
		s.mu.Lock()
		maps.Copy(s.metrics, extra)
		s.mu.Unlock()
	}
}

func (s *AgentService) postMetrics(metric model.Metrics) error {
	body, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	var sign string
	if s.config.SecretKey != "" {
		hmac := hmac.New(sha256.New, []byte(s.config.SecretKey))
		hmac.Write(body)
		sign = hex.EncodeToString(hmac.Sum(nil))
	}

	compressed, err := encoding.Compress(body)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("http://%s/update/", s.config.Address)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(compressed))
	if err != nil {
		return err
	}
	if sign != "" {
		req.Header.Set("HashSHA256", sign)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send bulk metrics: %s", resp.Status)
	}
	return nil
}

func (s *AgentService) pollLoop() {
	ticker := time.NewTicker(time.Duration(s.config.PollInterval) * time.Second)
	for range ticker.C {
		snapshot := s.collectRuntimeMetrics()
		s.mu.Lock()
		maps.Copy(s.metrics, snapshot)
		s.pollCount++
		s.mu.Unlock()
	}
}

func (s *AgentService) reportLoop(input chan model.Metrics) {
	ticker := time.NewTicker(time.Duration(s.config.ReportInterval) * time.Second)
	for range ticker.C {
		s.mu.Lock()
		snapshot := maps.Clone(s.metrics)
		pollCount := s.pollCount
		s.mu.Unlock()
		for name, value := range snapshot {
			input <- model.Metrics{
				ID:    name,
				MType: model.Gauge,
				Value: &value,
			}
		}
		input <- model.Metrics{
			ID:    "PollCount",
			MType: model.Counter,
			Delta: &pollCount,
		}
	}
}

func (s *AgentService) worker(input chan model.Metrics) {
	for metric := range input {
		if err := s.postMetrics(metric); err != nil {
			log.Printf("failed to send metric: %v\n", err)
		}
	}
}

func (s *AgentService) Run(quit chan os.Signal) {
	s.metrics = make(map[string]float64)
	jobs := make(chan model.Metrics, 40)

	workers := s.config.RateLimit

	if workers < 1 {
		workers = 1
	}

	go s.pollLoop()
	go s.additionalPollLoop()
	go s.reportLoop(jobs)

	for i := 0; i < workers; i++ {
		go s.worker(jobs)
	}

	<-quit
}
