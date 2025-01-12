package collector

import (
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
)

type MetricsCollector struct {
	storage interfaces.Storage
}

func NewCollector(s interfaces.Storage) *MetricsCollector {
	return &MetricsCollector{storage: s}
}

func (ms *MetricsCollector) StartCollectAndUpdate(pollInterval int) *time.Ticker {

	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)

	go func() {
		for range ticker.C {
			updateMetrics(ms.storage)
		}
	}()
	return ticker
}

func updateMetrics(memStorage interfaces.Storage) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memStorage.SetGauge("Alloc", float64(m.Alloc))
	memStorage.SetGauge("BuckHashSys", float64(m.BuckHashSys))
	memStorage.SetGauge("Frees", float64(m.Frees))
	memStorage.SetGauge("GCCPUFraction", float64(m.GCCPUFraction))
	memStorage.SetGauge("GCSys", float64(m.GCSys))
	memStorage.SetGauge("HeapAlloc", float64(m.HeapAlloc))
	memStorage.SetGauge("HeapIdle", float64(m.HeapIdle))
	memStorage.SetGauge("HeapInuse", float64(m.HeapInuse))
	memStorage.SetGauge("HeapObjects", float64(m.HeapObjects))
	memStorage.SetGauge("HeapReleased", float64(m.HeapReleased))
	memStorage.SetGauge("HeapSys", float64(m.HeapSys))
	memStorage.SetGauge("LastGC", float64(m.LastGC))
	memStorage.SetGauge("Lookups", float64(m.Lookups))
	memStorage.SetGauge("MCacheInuse", float64(m.MCacheInuse))
	memStorage.SetGauge("MSpanInuse", float64(m.MSpanInuse))
	memStorage.SetGauge("MSpanSys", float64(m.MSpanSys))
	memStorage.SetGauge("Mallocs", float64(m.Mallocs))
	memStorage.SetGauge("NextGC", float64(m.NextGC))
	memStorage.SetGauge("NumForcedGC", float64(m.NumForcedGC))
	memStorage.SetGauge("NumGC", float64(m.NumGC))
	memStorage.SetGauge("OtherSys", float64(m.OtherSys))
	memStorage.SetGauge("PauseTotalNs", float64(m.PauseTotalNs))
	memStorage.SetGauge("StackInuse", float64(m.StackInuse))
	memStorage.SetGauge("StackSys", float64(m.StackSys))
	memStorage.SetGauge("Sys", float64(m.Sys))
	memStorage.SetGauge("TotalAlloc", float64(m.TotalAlloc))

	memStorage.AddCounter("PollCount", 1)
	memStorage.SetGauge("RandomValue", rand.Float64())
}
