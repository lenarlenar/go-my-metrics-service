package main

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/middleware"
	"github.com/lenarlenar/go-my-metrics-service/internal/service"
	"github.com/lenarlenar/go-my-metrics-service/internal/storage"
)

var config flags.Config
func main() {
	var envConfig flags.EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		log.I().Fatalw(err.Error(), "event", "parse env")
	}

	storeInterval := *flag.Int("i", flags.DefaultStoreIntervalSec, "Интервал сохранения в файл")
	flag.StringVar(&config.ServerAddress, "a", flags.DefaultServerAddress , "Адрес сервера")
	flag.StringVar(&config.FileStoragePath, "f", flags.DefaultFileStoragePath, "Путь к файлу")
	flag.BoolVar(&config.Restore, "r", flags.DefaultRestore, "Загружать или нет ранее сохраненные файлы")
	flag.StringVar(&config.DatabaseDSN, "d", flags.DefaultDatabaseDSN, "Загружать или нет ранее сохраненные файлы")
	flag.Parse()

	if envConfig.ServerAddress != "" {
		config.ServerAddress = envConfig.ServerAddress
	}

	if _, isSet := os.LookupEnv("STORE_INTERVAL"); isSet {
		config.StoreInterval = time.Duration(envConfig.StoreInterval) * time.Second
	} else if storeInterval < 0 {
		config.StoreInterval = time.Duration(flags.DefaultStoreIntervalSec) * time.Second
	} else {
		config.StoreInterval = time.Duration(storeInterval) * time.Second
	}

	if envConfig.FileStoragePath != "" {
		config.FileStoragePath = envConfig.FileStoragePath
	}

	if _, isSet := os.LookupEnv("RESTORE"); isSet {
		config.Restore = envConfig.Restore
	}

	if envConfig.DatabaseDSN != "" {
		config.DatabaseDSN = envConfig.DatabaseDSN
	}

	storage := storage.NewStorage(config)
	//fileStore := filestore.FileStore{Storage: storage}
	//fileStore.Enable(fileStoragePath, storeInterval, restore)
	metricsService := service.NewService(storage)

	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.GzipCompression())
	router.Use(middleware.GzipUnpack())
	router.GET("/", metricsService.IndexHandler)
	router.GET("/ping", metricsService.PingHandler)
	router.POST("/value/", metricsService.ValueJSONHandler)
	router.POST("/update/", metricsService.UpdateJSONHandler)
	router.GET("/value/:type/:name/", metricsService.ValueHandler)
	router.POST("/update/:type/:name/:value", metricsService.UpdateHandler)

	log.I().Infoln(
		"Starting server",
		"addr", config.ServerAddress,
	)

	if err := router.Run(config.ServerAddress); err != nil {
		log.I().Fatalw(err.Error(), "event", "start server")
	}
}
