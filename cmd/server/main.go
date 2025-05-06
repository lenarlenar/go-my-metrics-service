package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/server/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/server/router"
	"github.com/lenarlenar/go-my-metrics-service/internal/service"
	"github.com/lenarlenar/go-my-metrics-service/internal/storage"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {

	log.I().Infof("Build version: %s\n", buildVersion)
	log.I().Infof("Build date: %s\n", buildDate)
	log.I().Infof("Build commit: %s\n", buildCommit)

	go func() {
		log.I().Infoln(
			"Starting debug/pprof server",
			"addr", "http://localhost:6060/debug/pprof/",
		)

		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.I().Fatalw(err.Error(), "event", "start debug/pprof server")
		}
	}()

	config := flags.Parse()
	storage := storage.NewStorage(config)
	metricsService := service.NewService(storage)
	router := router.New(config, metricsService)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.I().Infoln(
			"Starting server",
			"addr", config.ServerAddress,
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.I().Fatalw(err.Error(), "event", "start server")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.I().Fatalw(err.Error(), "event", "shutdown server")
	}
}
