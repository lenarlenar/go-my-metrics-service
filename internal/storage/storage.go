// Package storage содержит реализацию различных типов хранилищ метрик.
package storage

import (
	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/server/flags"
)

// NewStorage создает подходящее хранилище метрик в зависимости от конфигурации.
// Приоритет:
//  1. DBStorage — если указан DSN к базе данных,
//  2. FileStorage — если указан путь к файлу,
//  3. MemStorage — если ничего из вышеуказанного не задано или произошла ошибка при инициализации файла.
func NewStorage(config flags.Config) interfaces.Storage {
	if config.DatabaseDSN != "" {
		log.I().Info("тип хранилища: DBStorage")
		return NewDBStorage(config)
	} else if config.FileStoragePath != "" {
		fs, err := NewFileStorage(config)
		if err == nil {
			log.I().Info("тип хранилища: FileStorage")
			return fs
		}
	}

	log.I().Info("тип хранилища: MemStorage")
	return NewMemStorage()
}
