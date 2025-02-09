package storage

import (
	"github.com/lenarlenar/go-my-metrics-service/internal/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
)

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
