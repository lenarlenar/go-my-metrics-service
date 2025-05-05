package flags

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
)

const (
	defaultServerAddress  = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
	defaultKey            = ""
	defaultRateLimit      = 3
)

type EnvConfig struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

type Flags struct {
	ServerAddress  string
	PollInterval   time.Duration
	ReportInterval time.Duration
	Key            string
	RateLimit      int
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
	flag.Parse()

	if envConfig.ServerAddress != "" {
		*serverAddress = envConfig.ServerAddress
	}

	if envConfig.ReportInterval > 0 {
		*reportInterval = envConfig.ReportInterval
	} else if *reportInterval <= 0 {
		*reportInterval = 10
	}

	if envConfig.PollInterval > 0 {
		*pollInterval = envConfig.PollInterval
	} else if *pollInterval <= 0 {
		*pollInterval = 2
	}

	if envConfig.Key != "" {
		*key = envConfig.Key
	}

	if envConfig.RateLimit > 0 {
		*rateLimit = envConfig.RateLimit
	} else if *rateLimit < 0 {
		*rateLimit = 0
	}

	return Flags{
		ServerAddress:  *serverAddress,
		PollInterval:   time.Duration(*pollInterval) * time.Second,
		ReportInterval: time.Duration(*reportInterval) * time.Second,
		Key:            *key,
		RateLimit:      *rateLimit,
	}
}
