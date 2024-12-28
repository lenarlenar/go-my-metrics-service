package main

import (
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	Gauge map[string]float64
	Counter map[string]int64
}

type MetricsDB interface {
	SetGauge(n string, v float64)
	AddCounter(n string, v int64)
}

func (m *MemStorage) SetGauge(n string, v float64) {
	m.Gauge[n] = v
}

func (m *MemStorage) AddCounter(n string, v int64) {
	val, ok := m.Counter[n]
	if ok {
		m.Counter[n] = val + v
	} else {
		m.Counter[n] =  v
	}
}

var memStorage MemStorage
func main() {
	memStorage = MemStorage {map[string]float64{}, map[string]int64{}}
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, defaultHandler)
	return http.ListenAndServe(":8080", mux)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 5 {
		http.Error(w, "Invalid URL structure", http.StatusNotFound)
		return
	}

	metricType := parts[2]
	metricName := parts[3]

	switch metricType {
	case "gauge":
		if metricValue, err := strconv.ParseFloat(parts[4], 64); err != nil {
			http.Error(w, "Value must be float64", http.StatusBadRequest)
		} else {
			memStorage.SetGauge(metricName, metricValue)
		}
	case "counter":
		if metricValue, err := strconv.ParseInt(parts[4], 0, 64); err != nil {
			http.Error(w, "Value must be int64", http.StatusBadRequest)
		} else {
			memStorage.AddCounter(metricName, metricValue)
		}
	default:
		http.Error(w, "Unknown metric name", http.StatusBadRequest)
	}
}
