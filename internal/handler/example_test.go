package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Alexunder2003/alex-metrics-service/internal/config"
	"github.com/Alexunder2003/alex-metrics-service/internal/repository"
	"github.com/Alexunder2003/alex-metrics-service/internal/service"
	"go.uber.org/zap"
)

func exampleRouter() http.Handler {
	repo := repository.NewMemMetricsRepository()
	svc := service.NewMetricsService(repo, &config.ServerConfig{})
	return MetricsRouter(NewMetricsHandler(svc, zap.NewNop().Sugar()))
}

func ExampleHandler_UpdatePath() {
	// POST /update/{type}/{name}/{value} сохраняет counter и возвращает новое значение.
	srv := httptest.NewServer(exampleRouter())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/update/counter/PollCount/3", "text/plain", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.TrimSpace(string(body)))
	// Output:
	// 200
	// 3
}

func ExampleHandler_ValuePath() {
	// GET /value/{type}/{name} возвращает ранее сохранённую метрику текстом.
	srv := httptest.NewServer(exampleRouter())
	defer srv.Close()

	update, err := http.Post(srv.URL+"/update/gauge/Alloc/1.5", "text/plain", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	update.Body.Close()

	resp, err := http.Get(srv.URL + "/value/gauge/Alloc")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.TrimSpace(string(body)))
	// Output:
	// 200
	// 1.5
}

func ExampleHandler_UpdateJSON() {
	// POST /update/ принимает одну метрику JSON-объектом и возвращает её же.
	srv := httptest.NewServer(exampleRouter())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/update/", "application/json", strings.NewReader(
		`{"id":"Alloc","type":"gauge","value":1.5}`,
	))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.TrimSpace(string(body)))
	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":1.5}
}

func ExampleHandler_ValueJSON() {
	// POST /value/ принимает идентификатор метрики и возвращает сохранённое значение.
	srv := httptest.NewServer(exampleRouter())
	defer srv.Close()

	update, err := http.Post(srv.URL+"/update/", "application/json", strings.NewReader(
		`{"id":"PollCount","type":"counter","delta":7}`,
	))
	if err != nil {
		fmt.Println(err)
		return
	}
	update.Body.Close()

	resp, err := http.Post(srv.URL+"/value/", "application/json", strings.NewReader(
		`{"id":"PollCount","type":"counter"}`,
	))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.TrimSpace(string(body)))
	// Output:
	// 200
	// {"id":"PollCount","type":"counter","delta":7}
}

func ExampleHandler_BulkUpdateJSON() {
	// POST /updates/ сохраняет список метрик одним запросом.
	srv := httptest.NewServer(exampleRouter())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/updates/", "application/json", strings.NewReader(
		`[{"id":"PollCount","type":"counter","delta":2},{"id":"Alloc","type":"gauge","value":1.5}]`,
	))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.TrimSpace(string(body)))
	// Output:
	// 200
	// [{"id":"PollCount","type":"counter","delta":2},{"id":"Alloc","type":"gauge","value":1.5}]
}

func ExampleHandler_GetBulk() {
	// GET / отдаёт HTML-страницу со всеми сохранёнными метриками.
	srv := httptest.NewServer(exampleRouter())
	defer srv.Close()

	update, err := http.Post(srv.URL+"/update/counter/PollCount/3", "text/plain", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	update.Body.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	page := string(body)
	fmt.Println(resp.StatusCode)
	fmt.Println(strings.Contains(page, "PollCount"))
	fmt.Println(strings.Contains(page, "3"))
	// Output:
	// 200
	// true
	// true
}

func ExampleHealthHandler_DBHealthCheck() {
	// GET /ping проверяет базу. Без подключения обработчик отвечает 500.
	router := NewGlobalRouter(nil, []Mount{{
		Pattern: "/ping",
		Router:  HealthRouter(NewHealthHandler(nil, zap.NewNop().Sugar())),
	}})
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ping")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	// Output:
	// 500
}
