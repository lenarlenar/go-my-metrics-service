package service

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок для интерфейса Storage
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SetGauge(n string, v float64) {
	m.Called(n, v)
}

func (m *MockStorage) AddCounter(n string, v int64) {
	m.Called(n, v)
}

func (m *MockStorage) GetMetrics() map[string]model.Metrics {
	args := m.Called()
	return args.Get(0).(map[string]model.Metrics)
}

func (m *MockStorage) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func SetupRouter(s *MetricsService) *gin.Engine {
	r := gin.Default()
	r.GET("/ping", s.PingHandler)
	r.GET("/", s.IndexHandler)
	r.GET("/value/:type/:name/", s.ValueHandler)
	r.POST("/update/:type/:name/:value", s.UpdateHandler)
	r.POST("/value/", s.ValueJSONHandler)
	r.POST("/update/", s.UpdateJSONHandler)
	r.POST("/updates/", s.UpdateBatchHandler)
	return r
}

func TestPingHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("Ping").Return(nil)

	service := NewService(mockStorage)
	r := SetupRouter(service)

	w := performRequest(r, "GET", "/ping")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestIndexHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("GetMetrics").Return(map[string]model.Metrics{
		"metric1": {ID: "metric1", MType: "gauge", Value: float64Ptr(10.5)},
		"metric2": {ID: "metric2", MType: "counter", Delta: int64Ptr(20)},
	})

	service := NewService(mockStorage)
	r := SetupRouter(service)

	w := performRequest(r, "GET", "/")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "metric1")
	assert.Contains(t, w.Body.String(), "10.5")
	assert.Contains(t, w.Body.String(), "metric2")
	assert.Contains(t, w.Body.String(), "20")
}

func TestValueHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("GetMetrics").Return(map[string]model.Metrics{
		"metric1": {ID: "metric1", MType: "gauge", Value: float64Ptr(10.5)},
	})

	service := NewService(mockStorage)
	r := SetupRouter(service)

	w := performRequest(r, "GET", "/value/gauge/metric1/")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "10.5", w.Body.String())
}

func TestUpdateHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("SetGauge", "metric1", 20.5).Return(nil)

	service := NewService(mockStorage)
	r := SetupRouter(service)

	w := performRequest(r, "POST", "/update/gauge/metric1/20.5")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Запрос успешно обработан", w.Body.String())
}

func TestValueJSONHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("GetMetrics").Return(map[string]model.Metrics{
		"metric1": {ID: "metric1", MType: "gauge", Value: float64Ptr(10.5)},
	})

	service := NewService(mockStorage)
	r := SetupRouter(service)

	w := performRequest(r, "POST", "/value/", `{"id":"metric1"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"id":"metric1","type":"gauge","value":10.5}`, w.Body.String())
}

func TestUpdateJSONHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("SetGauge", "metric1", 20.5).Return(nil)
	mockStorage.On("GetMetrics").Return(map[string]model.Metrics{
		"metric1": {ID: "metric1", MType: "gauge", Value: float64Ptr(20.5)},
	})

	service := NewService(mockStorage)
	r := SetupRouter(service)

	w := performRequest(r, "POST", "/update/", `{"id":"metric1","type":"gauge","value":20.5}`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"id":"metric1","type":"gauge","value":20.5}`, w.Body.String())
}

func TestUpdateBatchHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	mockStorage.On("SetGauge", "metric1", 20.5).Return(nil)
	mockStorage.On("AddCounter", "metric2", int64(5)).Return(nil)

	service := NewService(mockStorage)
	r := SetupRouter(service)

	w := performRequest(r, "POST", "/updates/", `[{"id":"metric1","type":"gauge","value":20.5},{"id":"metric2","type":"counter","delta":5}]`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `"OK"`, w.Body.String())
}

func performRequest(r http.Handler, method, path string, body ...string) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if len(body) > 0 {
		reqBody = bytes.NewBufferString(body[0])
	}

	req := httptest.NewRequest(method, path, reqBody)
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
