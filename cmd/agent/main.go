package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"
	"github.com/lenarlenar/go-my-metrics-service/internal/db"
)

var memStorage db.MemStorage


func main() {

    urlGaugePattern := "http://localhost:8080/update/gauge/%s/%f"
    //urlCounterPattern := "http://localhost:8080/update/counter/%s/%f"
	memStorage = db.MemStorage{Gauge: map[string]float64{}, Counter: map[string]int64{}}
	tickerUpdateMetrics:= startUpdateRuntimeMetrics()
    defer tickerUpdateMetrics.Stop()
	i := 0

    for {
		memStorage.Mutex.Lock()
        for key, value := range memStorage.Gauge {
            url := fmt.Sprintf(urlGaugePattern, key, value)
			go sendPostRequest(url, i)
			i++
		}
		memStorage.Mutex.Unlock()
		time.Sleep(1 * time.Second)
    }
}

func startUpdateRuntimeMetrics() *time.Ticker  {

	ticker := time.NewTicker(2 * time.Second)

	go func() {
		for range ticker.C {
			updateRuntimeMetrics()
        }
	}()
    return ticker
}

func sendPostRequest(url string, i int) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("Ошибка при создании запроса: %v", err)
		return
	}

    req.Close = true
	req.Header.Set("Content-Length", "0")
	req.Header.Set("Content-Type", "text/plain")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка при отправке запроса: %d %v", i, err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Ответ от %s: %s\n", url, resp.Status)
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

	//memStorage.AddCounter("PollCount", 1)
	memStorage.SetGauge("RandomValue", rand.Float64())
}
