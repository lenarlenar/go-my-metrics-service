package flags

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
)

const (
	DefaultServerAddress    = "localhost:8080"
	DefaultStoreIntervalSec = 300
	DefaultFileStoragePath  = "" //"metrics.json"
	DefaultRestore          = true
	DefaultDatabaseDSN      = "" //"host=localhost port=5432 user=postgres password=admin dbname=postgres sslmode=disable"
	DefaultKey              = ""
)

type Config struct {
	ServerAddress   string        // адрес сервера, по умолчанию "localhost:8080"
	StoreInterval   time.Duration // интервал сохранения метрик в файл
	FileStoragePath string        // путь к файлу хранения метрик
	Restore         bool          // восстанавливать метрики из файла при старте
	DatabaseDSN     string        // строка подключения к БД PostgreSQL
	Key             string        // ключ для HMAC-подписи
}

type EnvConfig struct {
	ServerAddress   string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

func Parse() Config {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		log.I().Fatalw(err.Error(), "event", "parse env")
	}

	var config Config
	storeInterval := *flag.Int("i", DefaultStoreIntervalSec, "Интервал сохранения в файл")
	flag.StringVar(&config.ServerAddress, "a", DefaultServerAddress, "Адрес сервера")
	flag.StringVar(&config.FileStoragePath, "f", DefaultFileStoragePath, "Путь к файлу")
	flag.BoolVar(&config.Restore, "r", DefaultRestore, "Загружать или нет ранее сохраненные файлы")
	flag.StringVar(&config.DatabaseDSN, "d", DefaultDatabaseDSN, "Загружать или нет ранее сохраненные файлы")
	flag.StringVar(&config.Key, "k", DefaultKey, "Ключ для шифрования")
	flag.Parse()

	if envConfig.ServerAddress != "" {
		config.ServerAddress = envConfig.ServerAddress
	}

	if envConfig.Key != "" {
		config.Key = envConfig.Key
	}

	if _, isSet := os.LookupEnv("STORE_INTERVAL"); isSet {
		config.StoreInterval = time.Duration(envConfig.StoreInterval) * time.Second
	} else if storeInterval < 0 {
		config.StoreInterval = time.Duration(DefaultStoreIntervalSec) * time.Second
	} else {
		config.StoreInterval = time.Duration(storeInterval) * time.Second
	}

	if envConfig.FileStoragePath != "" {
		config.FileStoragePath = envConfig.FileStoragePath
	}

	if _, isSet := os.LookupEnv("RESTORE"); isSet {
		config.Restore = envConfig.Restore
	}

	if envConfig.DatabaseDSN != "" {
		config.DatabaseDSN = envConfig.DatabaseDSN
	}

	return config
}
