package sender

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lenarlenar/go-my-metrics-service/internal/agent/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
)

type MetricsSender struct {
	baseURL    string
	updateURL  string
	updatesURL string
	storage    interfaces.Storage
}

func NewSender(serverAddress string, memStorage interfaces.Storage) *MetricsSender {
	baseURL := fmt.Sprintf("http://%s", serverAddress)
	updateURL := fmt.Sprintf("%s/update/", baseURL)
	updatesURL := fmt.Sprintf("%s/updates/", baseURL)
	return &MetricsSender{baseURL: baseURL, updateURL: updateURL, updatesURL: updatesURL, storage: memStorage}
}

func (m *MetricsSender) Run(flags flags.Flags) {

	gzipIsSupported := gzipIsSupported(m.baseURL)
	log.I().Infof("Поддержка gzip: %v\n", gzipIsSupported)
	var rsaPub *rsa.PublicKey
	if flags.CryptoPath != "" {
		var err error
		rsaPub, err = loadPublicKey(flags.CryptoPath)
		if err != nil {
			log.I().Fatalf("ошибка загрузки RSA ключа: %v", err)
		}
	}
	for {
		// for _, model := range m.storage.GetMetrics() {
		// 	go sendPostRequest(m.updateURL, model)
		// 	go sendPostWithJSONRequest(m.updateURL, model, gzipIsSupported)
		// }
		go sendPostBatchRequest(flags.Key, m.updatesURL, m.storage.GetMetrics(), gzipIsSupported, rsaPub)
		time.Sleep(flags.ReportInterval)
	}
}

func Send(flags flags.Flags, metrics map[string]model.Metrics) {
	baseURL := fmt.Sprintf("http://%s", flags.ServerAddress)
	updatesURL := fmt.Sprintf("%s/updates/", baseURL)
	gzipIsSupported := gzipIsSupported(baseURL)
	var rsaPub *rsa.PublicKey
	if flags.CryptoPath != "" {
		var err error
		rsaPub, err = loadPublicKey(flags.CryptoPath)
		if err != nil {
			log.I().Fatalf("ошибка загрузки RSA ключа: %v", err)
		}
	}

	sendPostBatchRequest(flags.Key, updatesURL, metrics, gzipIsSupported, rsaPub)
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("неверный формат публичного ключа")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("это не RSA ключ")
	}

	return pubKey, nil
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzipWriterPool.Get().(*gzip.Writer)
	writer.Reset(&buf)

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	gzipWriterPool.Put(writer)

	return buf.Bytes(), nil
}

func gzipIsSupported(baseURL string) bool {
	resp, err := resty.New().R().
		SetHeader("Accept-Encoding", "gzip").
		Get(baseURL)

	if err != nil {
		log.I().Warnf("Не удалось проверить поддержку gzip: %v\n", err)
		return false
	}

	return resp.Header().Get("Content-Encoding") == "gzip"
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
		log.I().Warnf("ошибка при отправке запроса: %v", err)
		return
	}

	log.I().Infof("Ответ от %s: %d %s\n", url, resp.StatusCode(), resp)
}

func sendPostWithJSONRequest(url string, model model.Metrics, compress bool) {
	jsonModel, err := json.Marshal(model)
	if err != nil {
		log.I().Warnf("ошибка сериализатора: %v", err)
		return
	}
	client := resty.New()
	request := client.R().SetHeader("Content-Type", "application/json")
	if compress {
		request.SetHeader("Content-Encoding", "gzip")
		compressedData, err := compressData(jsonModel)
		if err != nil {
			log.I().Warnf("ошибка при попытке сжать метрику %s: %v\n", model.ID, err)
			return
		}
		request.SetBody(compressedData)
	} else {
		request.SetBody(jsonModel)
	}
	resp, err := request.Post(url)
	if err != nil {
		log.I().Warnf("ошибка при отправке запроса: %v", err)
		return
	}
	log.I().Infof("Ответ от %s: %d %s\n", url, resp.StatusCode(), resp)
}

func calculateHash(data, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func sendPostBatchRequest(
	key string,
	url string,
	metrics map[string]model.Metrics,
	compress bool,
	rsaPub *rsa.PublicKey,
) {
	metricsSlice := make([]model.Metrics, 0, len(metrics))
	for _, m := range metrics {
		metricsSlice = append(metricsSlice, m)
	}

	jsonModel, err := json.Marshal(metricsSlice)
	if err != nil {
		log.I().Warnf("ошибка сериализатора: %v", err)
		return
	}

	client := resty.New()
	request := client.R()

	var bodyToSend []byte

	switch {
	case compress && rsaPub == nil: //gzip только если нет RSA
		request.SetHeader("Content-Encoding", "gzip")
		request.SetHeader("Content-Type", "application/json")

		compressedData, err := compressData(jsonModel)
		if err != nil {
			log.I().Warnf("ошибка при сжатии: %v", err)
			return
		}
		bodyToSend = compressedData

	case rsaPub != nil: // шифрование RSA, без gzip
		request.SetHeader("Content-Type", "application/octet-stream")

		encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, jsonModel)
		if err != nil {
			log.I().Warnf("ошибка при шифровании: %v", err)
			return
		}
		bodyToSend = encrypted

	default:
		request.SetHeader("Content-Type", "application/json")
		bodyToSend = jsonModel
	}

	if key != "" && rsaPub == nil && !compress { //хеш только если нет RSA и gzip
		log.I().Info("secretKey: " + key)
		hash := calculateHash(jsonModel, []byte(key))
		log.I().Info("HashSHA256: " + hash)
		request.SetHeader("HashSHA256", hash)
	}

	request.SetBody(bodyToSend)

	resp, err := postWithRetry(request, url)
	if err != nil {
		log.I().Warnf("ошибка при отправке запроса: %v", err)
	} else {
		log.I().Infof("ответ от %s: %d %s\n", url, resp.StatusCode(), resp)
	}
}

const retryCount = 3

func postWithRetry(request *resty.Request, url string) (*resty.Response, error) {
	delay := 1
	for i := 0; i < retryCount; i++ {
		resp, err := request.Post(url)
		if err != nil {
			log.I().Warnf("ошибка при запросе к серверу: %v\n", err)
		} else if resp.StatusCode() == 200 {
			return resp, nil
		} else {
			log.I().Warnf("ошибка при запросе к серверу: status code %d\n", resp.StatusCode())
		}
		time.Sleep(time.Duration(delay) * time.Second)
		delay += 2
	}
	return nil, fmt.Errorf("запрос не удалось выполнить успешно после %d попыток", retryCount)
}
