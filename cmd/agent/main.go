package main

import (
	"sync"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/agent/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/collector"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
	"github.com/lenarlenar/go-my-metrics-service/internal/sender"
	"github.com/lenarlenar/go-my-metrics-service/internal/storage"
)

func main() {
	flags := flags.GetFlags()
	storage := storage.NewMemStorage()

	tickerPoll := time.NewTicker(flags.PollInterval)
	defer tickerPoll.Stop()
	go func() {
		for range tickerPoll.C {
			collector.UpdateMetrics(storage)
		}
	}()

	if flags.RateLimit == 0 {
		sender.NewSender(flags.ServerAddress, storage).Run(flags.ReportInterval, flags.Key)
	} else {

		go func() {
			for range tickerPoll.C {
				collector.UpdateExtraMetrics(storage)
			}
		}()

		mainChan := make(chan map[string]model.Metrics, flags.RateLimit)
		tickerReport := time.NewTicker(flags.ReportInterval)
		defer tickerReport.Stop()
		var wg sync.WaitGroup

		log.I().Info("Запускаем воркер...")
		for i := 0; i < flags.RateLimit; i++ {
			wg.Add(1)
			go worker(flags, mainChan, &wg)
		}
		log.I().Infof("Запушено воркеров: %d\n", flags.RateLimit)

		go func() {
			for range tickerReport.C {
				mainChan <- storage.GetMetrics()
			}
		}()

		select {}
	}
}

func worker(flags flags.Flags, channel chan map[string]model.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	for data := range channel {
		sender.Send(flags, data)
	}
}
