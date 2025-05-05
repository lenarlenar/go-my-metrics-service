// Package service содержит реализацию HTTP-хендлеров для работы с метриками.
package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

// MetricsService предоставляет методы для обработки запросов к метрикам.
type MetricsService struct {
	storage interfaces.Storage
}

// NewService создает новый экземпляр MetricsService с переданным хранилищем.
func NewService(s interfaces.Storage) *MetricsService {
	return &MetricsService{storage: s}
}

// PingHandler проверяет доступность хранилища и возвращает "pong", если всё ок.
func (s *MetricsService) PingHandler(c *gin.Context) {
	err := s.storage.Ping()
	if err != nil {
		log.I().Warnf("Ошибка при попытке вызова метода Ping к базе данных: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.String(http.StatusOK, "pong")
}

// IndexHandler возвращает HTML-страницу со списком всех метрик.
func (s *MetricsService) IndexHandler(c *gin.Context) {

	tableRows := ""

	for k, v := range s.storage.GetMetrics() {
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

// ValueHandler возвращает значение метрики по имени и типу из URL.
func (s *MetricsService) ValueHandler(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	if metric, ok := s.storage.GetMetrics()[metricName]; ok {
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

// UpdateHandler обновляет метрику по имени, типу и значению из URL.
func (s *MetricsService) UpdateHandler(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	switch metricType {
	case "gauge":
		if metricValue, err := strconv.ParseFloat(metricValue, 64); err != nil {
			c.String(http.StatusBadRequest, "Value must be float64")
		} else {
			s.storage.SetGauge(metricName, metricValue)
		}
	case "counter":
		if metricValue, err := strconv.ParseInt(metricValue, 0, 64); err != nil {
			c.String(http.StatusBadRequest, "Value must be int64")
		} else {
			s.storage.AddCounter(metricName, metricValue)
		}
	default:
		c.String(http.StatusBadRequest, "Unknown metric name")
	}

	c.String(http.StatusOK, "Запрос успешно обработан")
}

// ValueJSONHandler возвращает метрику по JSON-запросу.
func (s *MetricsService) ValueJSONHandler(c *gin.Context) {
	var requestMetric model.Metrics
	if err := c.ShouldBindJSON(&requestMetric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if metric, ok := s.storage.GetMetrics()[requestMetric.ID]; ok {
		c.JSON(http.StatusOK, metric)
	} else {
		c.JSON(http.StatusNotFound, "Unknown metric name")
	}
}

// UpdateJSONHandler обновляет метрику из JSON-запроса.
func (s *MetricsService) UpdateJSONHandler(c *gin.Context) {
	var metric model.Metrics
	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch metric.MType {
	case "gauge":
		s.storage.SetGauge(metric.ID, *metric.Value)
	case "counter":
		s.storage.AddCounter(metric.ID, *metric.Delta)
	default:
		c.JSON(http.StatusBadRequest, "Unknown metric name")
	}

	updatedMetric := s.storage.GetMetrics()[metric.ID]
	c.JSON(http.StatusOK, updatedMetric)
}

// UpdateBatchHandler обновляет сразу несколько метрик из JSON-массива.
func (s *MetricsService) UpdateBatchHandler(c *gin.Context) {
	metrics := make([]model.Metrics, 0)
	if err := c.ShouldBindJSON(&metrics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			s.storage.SetGauge(metric.ID, *metric.Value)
		case "counter":
			s.storage.AddCounter(metric.ID, *metric.Delta)
		default:
			log.I().Warnf("не известный тип метрики: %v", metric.MType)
			c.JSON(http.StatusBadRequest, "Unknown metric name")
		}
	}

	c.JSON(http.StatusOK, "OK")
}
