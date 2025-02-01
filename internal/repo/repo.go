package repo

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

type MemStorage struct {
	mutex   sync.Mutex
	metrics map[string]model.Metrics
}

func NewStorage() *MemStorage {
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

func (m *MemStorage) EnableFileBackup(filePath string, storeInterval int, restore bool) {

	if filePath == "" {
		log.I().Warn("Сохранение в файл не работает: невалидный путь.")
		return
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.I().Warnf("Ошибка при попытке открыть файл: %w", err)
		return
	}

	if restore {
		m.load(file)
	}

	go func() {
		for {
			time.Sleep(time.Duration(storeInterval) * time.Second)
			m.save(file)
		}
	}()
}

func (m *MemStorage) save(file *os.File) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := file.Truncate(0); err != nil {
		log.I().Errorf("Ошибка при попытке сохранить метрики в файл: %w", err)
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		log.I().Errorf("Ошибка при попытке сохранить метрики в файл: %w", err)
		return
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(m.metrics); err != nil {
		log.I().Errorf("Ошибка при попытке сохранить метрики в файл: %w", err)
		return
	}
}

func (m *MemStorage) load(file *os.File) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, err := file.Seek(0, 0); err != nil {
		log.I().Warnf("Ошибка при загрузке метрик с файла: %w", err)
		return
	}
	decoder := json.NewDecoder(file)
	var metrics map[string]model.Metrics
	for {
		if err := decoder.Decode(&metrics); err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				log.I().Warnf("Ошибка при загрузке метрик с файла: %w", err)
				return
			}
		}
		m.metrics = metrics
	}

}
