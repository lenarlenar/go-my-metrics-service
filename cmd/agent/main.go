package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
	"github.com/lenarlenar/go-my-metrics-service/internal/collector"
	"github.com/lenarlenar/go-my-metrics-service/internal/repo"
	"github.com/lenarlenar/go-my-metrics-service/internal/sender"
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
	storage := repo.NewStorage()
	metricsCollector := collector.NewCollector(storage)
	ticker := metricsCollector.StartCollectAndUpdate(flags.pollInterval)
	defer ticker.Stop()
	sender.NewSender(storage).Run(flags.reportInterval, flags.serverAddress)
}

func getFlags() Flags {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		log.Fatal(err)
	}

	serverAddress := flag.String("a", "localhost:8080", "HTTP server network address")
	reportInterval := flag.Int("r", 10, "Send data interval")
	pollInterval := flag.Int("p", 2, "Update data interval")
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
