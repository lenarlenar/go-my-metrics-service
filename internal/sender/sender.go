package sender

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

type MetricsSender struct {
	url     string
	storage interfaces.Storage
}

func NewSender(serverAddress string, memStorage interfaces.Storage) *MetricsSender {
	url := fmt.Sprintf("http://%s/update/", serverAddress)
	return &MetricsSender{url: url, storage: memStorage}
}

func (m *MetricsSender) Run(reportInterval int) {
	for {
		for _, model := range m.storage.GetMetrics() {
			go sendPostRequest(m.url, model)
			go sendPostWithJSONRequest(m.url, model)
		}
		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}

func sendPostRequest(url string, model model.Metrics) {
	var value string
	if model.MType == "gauge" {
		value = fmt.Sprintf("%g", *model.Value)
	} else {
		value = fmt.Sprintf("%d", *model.Delta)
	}

	fullURL := fmt.Sprintf("%s%s/%s/%s", url, model.MType, model.ID, value)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "text/plain").
		Post(fullURL)

	if err != nil {
		log.Printf("Ошибка при отправке запроса: %v", err)
		return
	}

	fmt.Printf("Ответ от %s: %d %s\n", url, resp.StatusCode(), resp)
}

func sendPostWithJSONRequest(url string, model model.Metrics) {

	jsonModel, err := json.Marshal(model)
	if err != nil {
		log.Printf("Ошибка при отправке запроса: %v", err)
		return
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(jsonModel).
		Post(url)

	if err != nil {
		log.Printf("Ошибка при отправке запроса: %v", err)
		return
	}

	fmt.Printf("Ответ от %s: %d %s\n", url, resp.StatusCode(), resp)
}
