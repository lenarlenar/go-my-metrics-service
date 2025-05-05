package model

type Metrics struct {
	ID    string   `json:"id"`              // Название метрики
	MType string   `json:"type"`            // Тип метрики: "gauge" или "counter"
	Delta *int64   `json:"delta,omitempty"` // Значение для counter (может быть nil)
	Value *float64 `json:"value,omitempty"` // Значение для gauge (может быть nil)
}
