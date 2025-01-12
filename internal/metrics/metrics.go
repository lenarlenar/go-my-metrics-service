package metrics

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
)

func ValueHandler(c *gin.Context, memStorage interfaces.MetricsDB) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	switch metricType {
	case "gauge":
		if val, ok := memStorage.GetGauge()[metricName]; ok {
			c.String(http.StatusOK, fmt.Sprintf("%g", val))
		} else {
			c.String(http.StatusNotFound, "Unknown metric name")
		}
	case "counter":
		if val, ok := memStorage.GetCounter()[metricName]; ok {
			c.String(http.StatusOK, fmt.Sprintf("%d", val))
		} else {
			c.String(http.StatusNotFound, "Unknown metric name")
		}
	default:
		c.String(http.StatusNotFound, "Unknown metric type")
	}
}

func UpdateHandler(c *gin.Context, memStorage interfaces.MetricsDB) {
	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	switch metricType {
	case "gauge":
		if metricValue, err := strconv.ParseFloat(metricValue, 64); err != nil {
			c.String(http.StatusBadRequest, "Value must be float64")
		} else {
			memStorage.SetGauge(metricName, metricValue)
		}
	case "counter":
		if metricValue, err := strconv.ParseInt(metricValue, 0, 64); err != nil {
			c.String(http.StatusBadRequest, "Value must be int64")
		} else {
			memStorage.AddCounter(metricName, metricValue)
		}
	default:
		c.String(http.StatusBadRequest, "Unknown metric name")
	}

	c.String(http.StatusOK, "Запрос успешно обработан")
}
