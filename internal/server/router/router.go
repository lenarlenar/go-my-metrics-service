// Package router отвечает за настройку маршрутов HTTP-сервера.
package router

import (
	"crypto/rsa"

	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/middleware"
	"github.com/lenarlenar/go-my-metrics-service/internal/server/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/service"
)

// New создает новый экземпляр роутера с зарегистрированными маршрутами и middleware.
func New(config flags.Config, metricsService *service.MetricsService, rsaKey *rsa.PrivateKey) *gin.Engine {
	router := gin.New()

	// Подключение middleware: логирование и gzip
	router.Use(middleware.Logger())
	router.Use(middleware.GzipCompression())
	router.Use(middleware.GzipUnpack())
	router.Use(middleware.TrustedSubnet(config.TrustedSubnet))

	// Группа роутов с проверкой подписи
	updatesGroup := router.Group("/updates")
	updatesGroup.Use(middleware.CheckHash(config.Key))
	updatesGroup.Use(middleware.RSADecrypt(rsaKey))
	{
		updatesGroup.POST("/", metricsService.UpdateBatchHandler)
	}

	// Общие маршруты
	router.GET("/", metricsService.IndexHandler)
	router.GET("/ping", metricsService.PingHandler)
	router.POST("/value/", metricsService.ValueJSONHandler)
	router.POST("/update/", metricsService.UpdateJSONHandler)
	router.GET("/value/:type/:name/", metricsService.ValueHandler)
	router.POST("/update/:type/:name/:value", metricsService.UpdateHandler)

	return router
}
