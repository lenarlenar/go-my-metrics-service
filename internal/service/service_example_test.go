package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

type fakeStorage struct {
	metrics map[string]model.Metrics
}

func (f *fakeStorage) Ping() error { return nil }

func (f *fakeStorage) GetMetrics() map[string]model.Metrics {
	return f.metrics
}

func (f *fakeStorage) SetGauge(name string, value float64) {
	f.metrics[name] = model.Metrics{ID: name, MType: "gauge", Value: &value}
}

func (f *fakeStorage) AddCounter(name string, delta int64) {
	if m, ok := f.metrics[name]; ok && m.Delta != nil {
		delta += *m.Delta
	}
	f.metrics[name] = model.Metrics{ID: name, MType: "counter", Delta: &delta}
}

// Пример PingHandler
func ExampleMetricsService_PingHandler() {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	s := NewService(&fakeStorage{metrics: make(map[string]model.Metrics)})
	router.GET("/ping", s.PingHandler)

	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
	// Output:
	// 200
	// pong
}

// Пример IndexHandler
func ExampleMetricsService_IndexHandler() {
	gin.SetMode(gin.TestMode)
	storage := &fakeStorage{
		metrics: map[string]model.Metrics{
			"temp":   {ID: "temp", MType: "gauge", Value: float64Ptr(42.5)},
			"visits": {ID: "visits", MType: "counter", Delta: int64Ptr(100)},
		},
	}
	s := NewService(storage)

	router := gin.New()
	router.GET("/", s.IndexHandler)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

// Пример ValueHandler
func ExampleMetricsService_ValueHandler() {
	gin.SetMode(gin.TestMode)
	storage := &fakeStorage{
		metrics: map[string]model.Metrics{
			"temp": {ID: "temp", MType: "gauge", Value: float64Ptr(42.5)},
		},
	}
	s := NewService(storage)
	router := gin.New()
	router.GET("/value/:type/:name", s.ValueHandler)

	req, _ := http.NewRequest("GET", "/value/gauge/temp", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
	// Output:
	// 200
	// 42.5
}

// Пример UpdateHandler
func ExampleMetricsService_UpdateHandler() {
	gin.SetMode(gin.TestMode)
	storage := &fakeStorage{metrics: make(map[string]model.Metrics)}
	s := NewService(storage)
	router := gin.New()
	router.POST("/update/:type/:name/:value", s.UpdateHandler)

	req, _ := http.NewRequest("POST", "/update/counter/test/5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
	// Output:
	// 200
	// Запрос успешно обработан
}

// Пример ValueJSONHandler
func ExampleMetricsService_ValueJSONHandler() {
	gin.SetMode(gin.TestMode)
	storage := &fakeStorage{
		metrics: map[string]model.Metrics{
			"cpu": {ID: "cpu", MType: "gauge", Value: float64Ptr(99.9)},
		},
	}
	s := NewService(storage)
	router := gin.New()
	router.POST("/value", s.ValueJSONHandler)

	body, _ := json.Marshal(model.Metrics{ID: "cpu", MType: "gauge"})
	req, _ := http.NewRequest("POST", "/value", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

// Пример UpdateJSONHandler
func ExampleMetricsService_UpdateJSONHandler() {
	gin.SetMode(gin.TestMode)
	storage := &fakeStorage{metrics: make(map[string]model.Metrics)}
	s := NewService(storage)
	router := gin.New()
	router.POST("/update", s.UpdateJSONHandler)

	val := 55.5
	metric := model.Metrics{ID: "load", MType: "gauge", Value: &val}
	body, _ := json.Marshal(metric)
	req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}

// Пример UpdateBatchHandler
func ExampleMetricsService_UpdateBatchHandler() {
	gin.SetMode(gin.TestMode)
	storage := &fakeStorage{metrics: make(map[string]model.Metrics)}
	s := NewService(storage)
	router := gin.New()
	router.POST("/updates", s.UpdateBatchHandler)

	metrics := []model.Metrics{
		{ID: "temp", MType: "gauge", Value: float64Ptr(23.5)},
		{ID: "clicks", MType: "counter", Delta: int64Ptr(10)},
	}
	body, _ := json.Marshal(metrics)
	req, _ := http.NewRequest("POST", "/updates", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output:
	// 200
}
