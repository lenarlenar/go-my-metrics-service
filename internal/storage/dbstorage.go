package storage

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/lenarlenar/go-my-metrics-service/internal/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

type DBStorage struct {
	mutex       sync.Mutex
	metrics     map[string]model.Metrics
	DB          sql.DB
	databaseDSN string
}

func NewDBStorage(config flags.Config) *DBStorage {
	return &DBStorage{
		metrics:     make(map[string]model.Metrics),
		databaseDSN: config.DatabaseDSN,
	}
}

func (m *DBStorage) GetMetrics() map[string]model.Metrics {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.metrics
}

func (m *DBStorage) SetMetrics(metrics map[string]model.Metrics) {
	m.mutex.Lock()
	m.metrics = metrics
	m.mutex.Unlock()
}

func (m *DBStorage) SetGauge(n string, v float64) {
	m.mutex.Lock()
	m.metrics[n] = model.Metrics{ID: n, MType: "gauge", Value: &v}
	m.mutex.Unlock()
}

func (m *DBStorage) AddCounter(n string, v int64) {
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

func (m *DBStorage) Ping() error {
	db, err := sql.Open("postgres", m.databaseDSN)
	if err != nil {
		return fmt.Errorf("ошибка при попытке подключиться к базе данных: %w", err)
	}

	return db.Ping()
}
