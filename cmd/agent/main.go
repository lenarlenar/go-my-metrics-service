package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/agent/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/collector"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/sender"
	"github.com/lenarlenar/go-my-metrics-service/internal/storage"
	"github.com/lenarlenar/go-my-metrics-service/internal/workerpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	flags := flags.GetFlags()
	storage := storage.NewMemStorage()

	tickerPoll := time.NewTicker(flags.PollInterval)
	defer tickerPoll.Stop()
	go func() {
		for {
			select {
			case <-tickerPoll.C:
				collector.UpdateMetrics(storage)
			case <-ctx.Done():
				return
			}
		}
	}()

	if flags.RateLimit == 0 {
		sender.NewSender(flags.ServerAddress, storage).Run(flags.ReportInterval, flags.Key)
	} else {
		go func() {
			for {
				select {
				case <-tickerPoll.C:
					collector.UpdateExtraMetrics(storage)
				case <-ctx.Done():
					return
				}
			}
		}()

		tickerReport := time.NewTicker(flags.ReportInterval)
		defer tickerReport.Stop()

		pool := workerpool.New(flags, flags.RateLimit)
		defer pool.Shutdown()

		log.I().Infof("Запушено воркеров: %d\n", flags.RateLimit)

		go func() {
			for {
				select {
				case <-tickerReport.C:
					metrics := storage.GetMetrics()
					if ok := pool.Submit(metrics); ok {
						log.I().Info("Метрики успешно отправлены в пул")
					} else {
						log.I().Warn("Пул переполнен, метрики не отправлены")
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		<- sigs
		log.I().Info("Получен сигнал завершения, останавливаемся...")
		cancel()
		log.I().Info("Агент завершил работу")
	}
}
