package sender

import (
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
)

type MetricsSender struct {
	urlPattern string
	storage    interfaces.Storage
}

func NewSender(memStorage interfaces.Storage) *MetricsSender {
	return &MetricsSender{urlPattern: "http://%s/update/%s/%s/%v", storage: memStorage}
}

func (m *MetricsSender) Run(reportInterval int, serverAddress string) {
	for {
		m.sendGauge(serverAddress)
		m.sendCounter(serverAddress)
		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}

func (m *MetricsSender) sendGauge(serverAddress string) {
	for key, value := range m.storage.GetGauge() {
		url := fmt.Sprintf(m.urlPattern, serverAddress, "gauge", key, value)
		go sendPostRequest(url)
	}
}

func (m *MetricsSender) sendCounter(serverAddress string) {
	for key, value := range m.storage.GetCounter() {
		url := fmt.Sprintf(m.urlPattern, serverAddress, "counter", key, value)
		go sendPostRequest(url)
	}
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
