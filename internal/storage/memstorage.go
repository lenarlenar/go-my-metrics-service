package storage

import (
	"errors"
	"sync"

	"github.com/lenarlenar/go-my-metrics-service/internal/model"
	_ "github.com/lib/pq"
)

type MemStorage struct {
	mutex   sync.Mutex
	metrics map[string]model.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]model.Metrics),
	}
}

func (m *MemStorage) GetMetrics() map[string]model.Metrics {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.metrics
}

func (m *MemStorage) SetGauge(n string, v float64) {
	m.mutex.Lock()
	m.metrics[n] = model.Metrics{ID: n, MType: "gauge", Value: &v}
	m.mutex.Unlock()
}

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

func (m *MemStorage) Ping() error {
	return errors.New("метод Ping() не определен для данного типа хранилища")
}
