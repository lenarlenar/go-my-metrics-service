package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/repo"
	"github.com/lenarlenar/go-my-metrics-service/internal/service"
)

var serverAddress string

type EnvConfig struct {
	ServerAddress string `env:"ADDRESS"`
}

func main() {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		log.Fatal(err)
	}

	if envConfig.ServerAddress == "" {
		flag.StringVar(&serverAddress, "a", "localhost:8080", "HTTP server network address")
		flag.Parse()
	} else {
		serverAddress = envConfig.ServerAddress
	}

	storage := repo.NewStorage()
	metricsService := service.NewService(storage)
	router := gin.Default()
	router.GET("/", metricsService.IndexHandler)
	router.GET("/value/:type/:name/", metricsService.ValueHandler)
	router.POST("/update/:type/:name/:value", metricsService.UpdateHandler)
	router.Run(serverAddress)
}
