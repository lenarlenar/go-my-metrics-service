package main

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/logger"
	"github.com/lenarlenar/go-my-metrics-service/internal/repo"
	"github.com/lenarlenar/go-my-metrics-service/internal/service"
	
)


var serverAddress string

type EnvConfig struct {
	ServerAddress string `env:"ADDRESS"`
}

func main() {


	localLogger := logger.GetLogger()
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		localLogger.Fatalw(err.Error(), "event", "parse env")
	}

	if envConfig.ServerAddress == "" {
		flag.StringVar(&serverAddress, "a", "localhost:8080", "HTTP server network address")
		flag.Parse()
	} else {
		serverAddress = envConfig.ServerAddress
	}

	storage := repo.NewStorage()
	metricsService := service.NewService(storage)
	
	router := gin.New()
	router.Use(localLogger.GetMiddleware())
	router.GET("/", metricsService.IndexHandler)
	router.GET("/value/:type/:name/", metricsService.ValueHandler)
	router.POST("/update/:type/:name/:value", metricsService.UpdateHandler)

	localLogger.Infoln(
		"Starting server",
		"addr", serverAddress,
	)

	if err := router.Run(serverAddress); err != nil {
		localLogger.Fatalw(err.Error(), "event", "start server")
	}
}
