package interfaces

import "github.com/lenarlenar/go-my-metrics-service/internal/model"

type Storage interface {
	SetGauge(n string, v float64)
	AddCounter(n string, v int64)
	GetMetrics() map[string]model.Metrics
	SetMetrics(m map[string]model.Metrics)
	Ping() error
}