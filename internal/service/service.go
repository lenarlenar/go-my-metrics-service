package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

type MetricsService struct {
	memStorage interfaces.Storage
}

func NewService(s interfaces.Storage) *MetricsService {
	return &MetricsService{memStorage: s}
}

func (s *MetricsService) IndexHandler(c *gin.Context) {

	tableRows := ""

	for k, v := range s.memStorage.GetMetrics() {
		switch v.MType {
		case "gauge":
			tableRows += "<tr><td>" + k + "</td><td>" + fmt.Sprintf("%g", *v.Value) + "</td></tr>"
		case "counter":
			tableRows += "<tr><td>" + k + "</td><td>" + fmt.Sprintf("%d", *v.Delta) + "</td></tr>"
		}
	}

	htmlContent := `
		<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Метрики</title>
				<style>
					table { border-collapse: collapse; width: 50%; margin: 20px auto; }
					th, td { border: 1px solid #ccc; padding: 8px; text-align: left; }
					th { background-color: #f4f4f4; }
				</style>
			</head>
			<body>
				<h1 style="text-align: center;">Метрики</h1>
				<table>
					<thead>
						<tr>
							<th>Метрика</th>
							<th>Значение</th>
						</tr>
					</thead>
					<tbody>
						` + tableRows + `
					</tbody>
				</table>
			</body>
			</html>
	`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

func (s *MetricsService) ValueHandler(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	if metric, ok := s.memStorage.GetMetrics()[metricName]; ok {
		var value string
		if metricType == "gauge" {
			value = fmt.Sprintf("%g", *metric.Value)
		} else {
			value = fmt.Sprintf("%d", *metric.Delta)
		}

		c.String(http.StatusOK, value)
	} else {
		c.String(http.StatusNotFound, "Unknown metric name")
	}
}

func (s *MetricsService) UpdateHandler(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	switch metricType {
	case "gauge":
		if metricValue, err := strconv.ParseFloat(metricValue, 64); err != nil {
			c.String(http.StatusBadRequest, "Value must be float64")
		} else {
			s.memStorage.SetGauge(metricName, metricValue)
		}
	case "counter":
		if metricValue, err := strconv.ParseInt(metricValue, 0, 64); err != nil {
			c.String(http.StatusBadRequest, "Value must be int64")
		} else {
			s.memStorage.AddCounter(metricName, metricValue)
		}
	default:
		c.String(http.StatusBadRequest, "Unknown metric name")
	}

	c.String(http.StatusOK, "Запрос успешно обработан")
}

func (s *MetricsService) ValueJSONHandler(c *gin.Context) {
	var requestMetric model.Metrics
	if err := c.ShouldBindJSON(&requestMetric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if metric, ok := s.memStorage.GetMetrics()[requestMetric.ID]; ok {
		c.JSON(http.StatusOK, metric)
	} else {
		c.JSON(http.StatusNotFound, "Unknown metric name")
	}
}

func (s *MetricsService) UpdateJSONHandler(c *gin.Context) {
	var metric model.Metrics
	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch metric.MType {
	case "gauge":
		s.memStorage.SetGauge(metric.ID, *metric.Value)
	case "counter":
		s.memStorage.AddCounter(metric.ID, *metric.Delta)
	default:
		c.JSON(http.StatusBadRequest, "Unknown metric name")
	}

	updatedMetric := s.memStorage.GetMetrics()[metric.ID]
	c.JSON(http.StatusOK, updatedMetric)
}
