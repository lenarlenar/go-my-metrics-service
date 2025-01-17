package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/caarlos0/env"
	"github.com/go-resty/resty/v2"
	"github.com/lenarlenar/go-my-metrics-service/internal/db"
)

var memStorage db.MemStorage
var reportInterval int
var pollInterval int
var serverAddress string

type EnvConfig struct {
    ServerAddress string `env:"ADDRESS"`
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval int `env:"POLL_INTERVAL"`
}

func main() {

	var envConfig EnvConfig
    if err := env.Parse(&envConfig); err != nil {
        log.Fatal(err)
    }

	flag.StringVar(&serverAddress, "a", "localhost:8080", "HTTP server network address")
	flag.IntVar(&reportInterval, "r", 10, "reportInterval")
	flag.IntVar(&pollInterval, "p", 2, "pollInterval")
	flag.Parse()

	if envConfig.ServerAddress != "" {
		serverAddress = envConfig.ServerAddress
	}

	if envConfig.PollInterval != 0 {
		pollInterval = envConfig.PollInterval
	}

	if envConfig.ReportInterval != 0 {
		reportInterval = envConfig.ReportInterval
	}

    urlGaugePattern := "http://%s/update/gauge/%s/%f"
    urlCounterPattern := "http://%s/update/counter/%s/%d"
	memStorage = db.MemStorage{Gauge: map[string]float64{}, Counter: map[string]int64{}}
	tickerUpdateMetrics:= startUpdateRuntimeMetrics()
    defer tickerUpdateMetrics.Stop()

    for {
		memStorage.Mutex.Lock()
        for key, value := range memStorage.Gauge {
            url := fmt.Sprintf(urlGaugePattern, serverAddress, key, value)
			go sendPostRequest(url)
		}
		for key, value := range memStorage.Counter {
            url := fmt.Sprintf(urlCounterPattern, serverAddress, key, value)
			go sendPostRequest(url)
		}
		memStorage.Mutex.Unlock()
		time.Sleep(time.Duration(reportInterval) * time.Second)
    }
}

func startUpdateRuntimeMetrics() *time.Ticker  {

	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)

	go func() {
		for range ticker.C {
			updateRuntimeMetrics()
        }
	}()
    return ticker
}

func sendPostRequest(url string) {

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "text/plain").
        Post(url)

	if err != nil {
		log.Printf("Ошибка при отправке запроса: %v", err)
		return
	}

	fmt.Printf("Ответ от %s: %d %s\n", url, resp.StatusCode(), resp)
}

func updateRuntimeMetrics() {
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
