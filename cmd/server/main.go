package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/db"
	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
	"github.com/lenarlenar/go-my-metrics-service/internal/metrics"
)

var memStorage interfaces.MetricsDB
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

	memStorage = db.NewMemStorage()
	router := gin.Default()
	router.POST("/update/:type/:name/:value", func(ctx *gin.Context) {
		metrics.UpdateHandler(ctx, memStorage)
	})
	router.GET("/value/:type/:name/", func(ctx *gin.Context) {
		metrics.ValueHandler(ctx, memStorage)
	})
	router.Run(serverAddress)
}