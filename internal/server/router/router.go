package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/middleware"
	"github.com/lenarlenar/go-my-metrics-service/internal/server/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/service"
)

func New(config flags.Config, metricsService *service.MetricsService) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.GzipCompression())
	router.Use(middleware.GzipUnpack())

	updatesGroup := router.Group("/updates")
	updatesGroup.Use(middleware.CheckHash(config.Key))
	{
		updatesGroup.POST("/", metricsService.UpdateBatchHandler)
	}
	router.GET("/", metricsService.IndexHandler)
	router.GET("/ping", metricsService.PingHandler)
	router.POST("/value/", metricsService.ValueJSONHandler)
	router.POST("/update/", metricsService.UpdateJSONHandler)
	router.GET("/value/:type/:name/", metricsService.ValueHandler)
	router.POST("/update/:type/:name/:value", metricsService.UpdateHandler)

	return router
}
