package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/db"
)

var memStorage db.MemStorage

func main() {
	memStorage = db.MemStorage{Gauge: map[string]float64{}, Counter: map[string]int64{}}
	if err := runServer(); err != nil {
		panic(err)
	}
}

func runServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, defaultHandler)

	server := &http.Server{
        Addr:           ":8080",
        Handler:        mux,
        ReadTimeout:    5 * time.Second,
        WriteTimeout:   10 * time.Second,
        IdleTimeout:    15 * time.Second,
    }

	return server.ListenAndServe()
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

	w.Write([]byte("Запрос успешно обработан"))
}
