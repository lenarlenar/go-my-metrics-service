package filestore

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

type FileStore struct {
	Storage interfaces.Storage
}

func (fs *FileStore) Enable(filePath string, storeInterval int, restore bool) {

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
		fs.load(file)
	}

	go func() {
		for {
			time.Sleep(time.Duration(storeInterval) * time.Second)
			fs.save(file)
		}
	}()
}

func (fs *FileStore) save(file *os.File) {
	if err := file.Truncate(0); err != nil {
		log.I().Errorf("Ошибка при попытке сохранить метрики в файл: %w", err)
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		log.I().Errorf("Ошибка при попытке сохранить метрики в файл: %w", err)
		return
	}

	buf := bufio.NewWriter(file)
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(fs.Storage.GetMetrics()); err != nil {
		log.I().Errorf("Ошибка при попытке сохранить метрики в файл: %w", err)
		return
	}
	buf.Flush()
}

func (fs *FileStore) load(file *os.File) {
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
		fs.Storage.SetMetrics(metrics) 
	}
}