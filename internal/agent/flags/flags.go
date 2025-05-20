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
	defaultConfigPath     = ""
	defaultServerAddress  = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
	defaultKey            = ""
	defaultRateLimit      = 3
	defaultCryptoPath     = ""
)

type JSONConfig struct {
	ServerAddress  string `json:"address"`
	ReportInterval int    `json:"report_interval"`
	PollInterval   int    `json:"poll_interval"`
	Key            string `json:"key"`
	RateLimit      int    `json:"rate_limit"`
	CryptoPath     string `json:"crypto_key"`
}

type EnvConfig struct {
	ConfigPath     string `env:"CONFIG"`
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoPath     string `env:"CRYPTO_KEY"`
}

type Flags struct {
	ServerAddress  string
	PollInterval   time.Duration
	ReportInterval time.Duration
	Key            string
	RateLimit      int
	CryptoPath     string
}

func GetFlags() Flags {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		log.I().Fatal(err)
	}

	serverAddress := flag.String("a", defaultServerAddress, "Адрес сервера")
	reportInterval := flag.Int("r", defaultReportInterval, "Интервал отправки на сервер")
	pollInterval := flag.Int("p", defaultPollInterval, "Интервал локального обновления данных")
	key := flag.String("k", defaultKey, "Ключ для шифрования")
	rateLimit := flag.Int("l", defaultRateLimit, "Количество одновременно исходящих запросов на сервер")
	cryptoPath := flag.String("crypto-key", defaultCryptoPath, "Путь до файла с приватным ключом")
	configPath := flag.String("c", defaultConfigPath, "Путь к конфиг-файлу JSON")
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

	return Flags{
		ServerAddress: coalesceString(
			envConfig.ServerAddress,
			*serverAddress,
			jsonConfig.ServerAddress,
			defaultServerAddress,
		),
		ReportInterval: time.Duration(coalesceInt(
			envConfig.ReportInterval,
			*reportInterval,
			jsonConfig.ReportInterval,
			defaultReportInterval,
		)) * time.Second,
		PollInterval: time.Duration(coalesceInt(
			envConfig.PollInterval,
			*pollInterval,
			jsonConfig.PollInterval,
			defaultPollInterval,
		)) * time.Second,
		Key: coalesceString(
			envConfig.Key,
			*key,
			jsonConfig.Key,
			defaultKey,
		),
		RateLimit: coalesceInt(
			envConfig.RateLimit,
			*rateLimit,
			jsonConfig.RateLimit,
			defaultRateLimit,
		),
		CryptoPath: coalesceString(
			envConfig.CryptoPath,
			*cryptoPath,
			jsonConfig.CryptoPath,
			defaultCryptoPath,
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
