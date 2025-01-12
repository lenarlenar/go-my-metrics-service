package interfaces

type Storage interface {
	SetGauge(n string, v float64)
	AddCounter(n string, v int64)
	GetGauge() map[string]float64
	GetCounter() map[string]int64
}
