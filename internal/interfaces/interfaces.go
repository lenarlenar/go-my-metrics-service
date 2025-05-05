package interfaces

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

type Collector interface {
	StartCollectAndUpdate(pollInterval int) *time.Ticker
}

type Storage interface {
	SetGauge(n string, v float64)
	AddCounter(n string, v int64)
	GetMetrics() map[string]model.Metrics
	Ping() error
}

type Service interface {
	IndexHandler(c *gin.Context)
	ValueHandler(c *gin.Context)
	UpdateHandler(c *gin.Context)
	ValueJSONHandler(c *gin.Context)
	UpdateJSONHandler(c *gin.Context)
}

type Sender interface {
	Run(reportInterval int, serverAddress string)
}
