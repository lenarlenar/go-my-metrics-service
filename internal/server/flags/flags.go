package flags

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
)

const (
	DefaultConfigPath       = ""
	DefaultServerAddress    = "localhost:8080"
	DefaultStoreIntervalSec = 300
	DefaultFileStoragePath  = "" //"metrics.json"
	DefaultRestore          = true
	DefaultDatabaseDSN      = "" //"host=localhost port=5432 user=postgres password=admin dbname=postgres sslmode=disable"
	DefaultKey              = ""
	DefaultCryptoPath       = ""
)

type JSONConfig struct {
	ServerAddress   string `json:"address"`
	StoreInterval   int    `json:"store_interval"`
	FileStoragePath string `json:"store_file"`
	Restore         bool   `json:"restore"`
	DatabaseDSN     string `json:"database_dsn"`
	CryptoPath      string `json:"crypto_key"`
}

type Config struct {
	ServerAddress   string        // адрес сервера, по умолчанию "localhost:8080"
	StoreInterval   time.Duration // интервал сохранения метрик в файл
	FileStoragePath string        // путь к файлу хранения метрик
	Restore         bool          // восстанавливать метрики из файла при старте
	DatabaseDSN     string        // строка подключения к БД PostgreSQL
	Key             string        // ключ для HMAC-подписи
	CryptoPath      string        // путь до файла с приватным ключом
}

type EnvConfig struct {
	ConfigPath      string `env:"CONFIG"`
	ServerAddress   string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	CryptoPath      string `env:"CRYPTO_KEY"`
}

func Parse() Config {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		log.I().Fatalw(err.Error(), "event", "parse env")
	}

	storeInterval := flag.Int("i", DefaultStoreIntervalSec, "Интервал сохранения в файл")
	serverAddress := flag.String("a", DefaultServerAddress, "Адрес сервера")
	fileStoragePath := flag.String("f", DefaultFileStoragePath, "Путь к файлу")
	restore := flag.Bool("r", DefaultRestore, "Загружать или нет ранее сохраненные файлы")
	databaseDSN := flag.String("d", DefaultDatabaseDSN, "Загружать или нет ранее сохраненные файлы")
	key := flag.String("k", DefaultKey, "Ключ для шифрования")
	cryptoPath := flag.String("crypto-key", DefaultCryptoPath, "Путь до файла с приватным ключом")
	configPath := flag.String("c", DefaultConfigPath, "Путь до файла с приватным ключом")
	flag.Parse()

	jsonConfig := &JSONConfig{}
	if envConfig.ConfigPath != "" {
		*configPath = envConfig.ConfigPath
	}
	if *configPath != "" {
		cfg, err := loadJSONConfig(*configPath)
		if err == nil {
			jsonConfig = cfg
		}
	}

	return Config{
		ServerAddress: coalesceString(
			envConfig.ServerAddress,
			*serverAddress,
			func() string { //иначе линтер ругается
				if jsonConfig != nil {
					return jsonConfig.ServerAddress
				}
				return ""
			}(),
			DefaultServerAddress,
		),
		StoreInterval: time.Duration(coalesceInt(
			envConfig.StoreInterval,
			*storeInterval,
			jsonConfig.StoreInterval,
			DefaultStoreIntervalSec,
		)) * time.Second,
		FileStoragePath: coalesceString(
			envConfig.FileStoragePath,
			*fileStoragePath,
			jsonConfig.FileStoragePath,
			DefaultFileStoragePath,
		),
		Restore: coalesceBoolPtr(
			lookupBoolEnv("RESTORE", envConfig.Restore),
			flagIsPassed("r", *restore),
			jsonConfig != nil && jsonConfig.Restore,
			DefaultRestore,
		),
		DatabaseDSN: coalesceString(
			envConfig.DatabaseDSN,
			*databaseDSN,
			jsonConfig.DatabaseDSN,
			DefaultDatabaseDSN,
		),
		Key: coalesceString(
			envConfig.Key,
			*key,
			DefaultKey,
			DefaultKey,
		),
		CryptoPath: coalesceString(
			envConfig.CryptoPath,
			*cryptoPath,
			jsonConfig.CryptoPath,
			DefaultCryptoPath,
		),
	}
}

func loadJSONConfig(path string) (*JSONConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg JSONConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func coalesceString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func coalesceInt(values ...int) int {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}

func coalesceBoolPtr(values ...bool) bool {
	for _, v := range values {
		return v
	}
	return false
}

func lookupBoolEnv(name string, val bool) bool {
	_, found := os.LookupEnv(name)
	if found {
		return val
	}
	return false
}

func flagIsPassed(name string, val bool) bool {
	visited := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			visited = true
		}
	})
	if visited {
		return val
	}
	return false
}
