package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/agent/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/collector"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
	"github.com/lenarlenar/go-my-metrics-service/internal/sender"
	"github.com/lenarlenar/go-my-metrics-service/internal/storage"
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
			for {
				select {
				case <-tickerReport.C:
					select {
					case mainChan <- storage.GetMetrics():
						log.I().Info("Метрики успешно отправлены в канал")
					default:
						log.I().Warn("Канал переполнен, метрики не отправлены")
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		<- sigs
		log.I().Info("Получен сигнал завершения, останавливаемся...")
		cancel()
		close(mainChan)
	 	wg.Wait()
		log.I().Info("Агент завершил работу")
	}
}

func worker(flags flags.Flags, channel chan map[string]model.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	for data := range channel {
		sender.Send(flags, data)
	}
}
