package main

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/middleware"
	"github.com/lenarlenar/go-my-metrics-service/internal/repo"
	"github.com/lenarlenar/go-my-metrics-service/internal/service"
)

const (
	defaultServerAddress   = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "metrics.json"
	defaultRestore         = true
)

var (
	serverAddress   string
	storeInterval   int
	fileStoragePath string
	restore         bool
)

type EnvConfig struct {
	ServerAddress   string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func main() {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		log.I().Fatalw(err.Error(), "event", "parse env")
	}

	flag.StringVar(&serverAddress, "a", defaultServerAddress, "Адрес сервера")
	flag.IntVar(&storeInterval, "i", defaultStoreInterval, "Интервал сохранения в файл")
	flag.StringVar(&fileStoragePath, "f", defaultFileStoragePath, "Путь к файлу")
	flag.BoolVar(&restore, "r", defaultRestore, "Загружать или нет ранее сохраненные файлы")
	flag.Parse()

	if envConfig.ServerAddress != "" {
		serverAddress = envConfig.ServerAddress
	}

	if _, isSet := os.LookupEnv("STORE_INTERVAL"); isSet {
		storeInterval = envConfig.StoreInterval
	} else if storeInterval < 0 {
		storeInterval = defaultStoreInterval
	}

	if envConfig.FileStoragePath != "" {
		fileStoragePath = envConfig.FileStoragePath
	}

	if _, isSet := os.LookupEnv("RESTORE"); isSet {
		restore = envConfig.Restore
	}

	storage := repo.NewStorage()
	storage.EnableFileBackup(fileStoragePath, storeInterval, restore)
	metricsService := service.NewService(storage)

	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.GzipCompression())
	router.Use(middleware.GzipUnpack())
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
