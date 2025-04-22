// Package storage предоставляет реализацию in-memory хранилища метрик.
package storage

import (
	"errors"
	"sync"

	"github.com/lenarlenar/go-my-metrics-service/internal/model"
	_ "github.com/lib/pq"
)

// Package storage предоставляет реализацию in-memory хранилища метрик.
type MemStorage struct {
	mutex   sync.Mutex
	metrics map[string]model.Metrics
}

// NewMemStorage создает новый экземпляр in-memory хранилища.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]model.Metrics),
	}
}

// GetMetrics возвращает копию всех метрик из памяти.
func (m *MemStorage) GetMetrics() map[string]model.Metrics {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.metrics
}

// SetGauge устанавливает значение метрики типа gauge.
func (m *MemStorage) SetGauge(n string, v float64) {
	m.mutex.Lock()
	m.metrics[n] = model.Metrics{ID: n, MType: "gauge", Value: &v}
	m.mutex.Unlock()
}

// AddCounter увеличивает значение метрики типа counter на заданную величину.
// Если метрика отсутствует — она создается.
func (m *MemStorage) AddCounter(n string, v int64) {
	m.mutex.Lock()
	oldMetric, ok := m.metrics[n]
	if ok {
		newDelta := *oldMetric.Delta + v
		updatedMetric := model.Metrics{ID: n, MType: "counter", Delta: &newDelta}
		m.metrics[n] = updatedMetric
	} else {
		m.metrics[n] = model.Metrics{ID: n, MType: "counter", Delta: &v}
	}
	m.mutex.Unlock()
}

// Ping возвращает ошибку, так как MemStorage не поддерживает подключение.
func (m *MemStorage) Ping() error {
	return errors.New("метод Ping() не определен для данного типа хранилища")
}
