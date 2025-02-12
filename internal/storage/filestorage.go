package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

type FileStorage struct {
	mutex   sync.Mutex
	metrics map[string]model.Metrics
}

func NewFileStorage(config flags.Config) (*FileStorage, error) {

	fs := &FileStorage{
		metrics: make(map[string]model.Metrics),
	}

	file, err := os.OpenFile(config.FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("ошибка при попытке открыть файл: %w", err)
	}

	if config.Restore {
		fs.load(file)
	}

	go func() {
		for {
			time.Sleep(config.StoreInterval)
			fs.save(file)
		}
	}()

	return fs, nil
}

func (fs *FileStorage) GetMetrics() map[string]model.Metrics {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	return fs.metrics
}

func (fs *FileStorage) SetGauge(n string, v float64) {
	fs.mutex.Lock()
	fs.metrics[n] = model.Metrics{ID: n, MType: "gauge", Value: &v}
	fs.mutex.Unlock()
}

func (fs *FileStorage) AddCounter(n string, v int64) {
	fs.mutex.Lock()
	oldMetric, ok := fs.metrics[n]
	if ok {
		newDelta := *oldMetric.Delta + v
		updatedMetric := model.Metrics{ID: n, MType: "counter", Delta: &newDelta}
		fs.metrics[n] = updatedMetric
	} else {
		fs.metrics[n] = model.Metrics{ID: n, MType: "counter", Delta: &v}
	}
	fs.mutex.Unlock()
}

func (fs *FileStorage) Ping() error {
	return errors.New("метод Ping() не определен для данного типа хранилища")
}

func (fs *FileStorage) save(file *os.File) {
	if err := file.Truncate(0); err != nil {
		log.I().Errorf("ошибка при попытке сохранить метрики в файл: %w", err)
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		log.I().Errorf("ошибка при попытке сохранить метрики в файл: %w", err)
		return
	}

	buf := bufio.NewWriter(file)
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(fs.GetMetrics()); err != nil {
		log.I().Errorf("ошибка при попытке сохранить метрики в файл: %w", err)
		return
	}
	buf.Flush()
}

func (fs *FileStorage) load(file *os.File) {
	if _, err := file.Seek(0, 0); err != nil {
		log.I().Warnf("ошибка при загрузке метрик с файла: %w", err)
		return
	}
	decoder := json.NewDecoder(file)
	var metrics map[string]model.Metrics
	for {
		if err := decoder.Decode(&metrics); err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				log.I().Warnf("ошибка при загрузке метрик с файла: %w", err)
				return
			}
		}
		fs.mutex.Lock()
		fs.metrics = metrics
		fs.mutex.Unlock()
	}
}
