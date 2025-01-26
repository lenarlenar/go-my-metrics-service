package main

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/middleware"
	"github.com/lenarlenar/go-my-metrics-service/internal/repo"
	"github.com/lenarlenar/go-my-metrics-service/internal/service"
)

const defaultServerAddress = "localhost:8080"

var serverAddress string

type EnvConfig struct {
	ServerAddress string `env:"ADDRESS"`
}

func main() {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		log.I().Fatalw(err.Error(), "event", "parse env")
	}

	if envConfig.ServerAddress == "" {
		flag.StringVar(&serverAddress, "a", defaultServerAddress, "HTTP server network address")
		flag.Parse()
	} else {
		serverAddress = envConfig.ServerAddress
	}

	storage := repo.NewStorage()
	metricsService := service.NewService(storage)

	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.GzipCompression())
	router.GET("/", metricsService.IndexHandler)
	router.POST("/value/", metricsService.ValueJSONHandler)
	router.POST("/update/", metricsService.UpdateJSONHandler)
	router.GET("/value/:type/:name/", metricsService.ValueHandler)
	router.POST("/update/:type/:name/:value", metricsService.UpdateHandler)

	log.I().Infoln(
		"Starting server",
		"addr", serverAddress,
	)

	if err := router.Run(serverAddress); err != nil {
		log.I().Fatalw(err.Error(), "event", "start server")
	}
}
