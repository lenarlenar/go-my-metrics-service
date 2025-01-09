package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/db"
)

var memStorage db.MemStorage
var serverAddress string

func main() {

	flag.StringVar(&serverAddress, "a", "localhost:8080", "HTTP server network address")
	flag.Parse()

	memStorage = db.MemStorage{Gauge: map[string]float64{}, Counter: map[string]int64{}}
	router := gin.Default()
	router.POST("/update/:type/:name/:value", updateHandler)
	router.GET("/value/:type/:name/", valueHandler)
	router.Run(serverAddress)
}

func valueHandler(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	switch metricType {
	case "gauge":
		if val, ok := memStorage.Gauge[metricName]; ok {
			c.String(http.StatusOK, fmt.Sprintf("%g", val))
		} else {
			c.String(http.StatusNotFound, "Unknown metric name")
		}
	case "counter":
		if val, ok := memStorage.Counter[metricName]; ok {
			c.String(http.StatusOK, fmt.Sprintf("%d", val))
		} else {
			c.String(http.StatusNotFound, "Unknown metric name")	
		}
	default:
		c.String(http.StatusNotFound, "Unknown metric type")
	}
}

func updateHandler(c *gin.Context) {
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