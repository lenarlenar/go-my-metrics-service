package main

import (
	"flag"

	"github.com/caarlos0/env"
	"github.com/lenarlenar/go-my-metrics-service/internal/collector"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/sender"
	"github.com/lenarlenar/go-my-metrics-service/internal/storage"
)

const (
	defaultServerAddress  = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
)

type EnvConfig struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

type Flags struct {
	serverAddress  string
	pollInterval   int
	reportInterval int
}

func main() {
	flags := getFlags()
	storage := storage.NewMemStorage()
	metricsCollector := collector.NewCollector(storage)
	ticker := metricsCollector.StartCollectAndUpdate(flags.pollInterval)
	defer ticker.Stop()
	sender.NewSender(flags.serverAddress, storage).Run(flags.reportInterval)
}

func getFlags() Flags {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		log.I().Fatal(err)
	}

	serverAddress := flag.String("a", defaultServerAddress, "Адрес сервера")
	reportInterval := flag.Int("r", defaultReportInterval, "Интервал отправки на сервер")
	pollInterval := flag.Int("p", defaultPollInterval, "Интервал локального обновления данных")
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

	return Flags{
		serverAddress:  *serverAddress,
		pollInterval:   *pollInterval,
		reportInterval: *reportInterval,
	}
}
