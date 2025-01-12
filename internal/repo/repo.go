package repo

import "sync"

type MemStorage struct {
	Mutex   sync.Mutex
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (m *MemStorage) GetCounter() map[string]int64 {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.Counter
}

func (m *MemStorage) GetGauge() map[string]float64 {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	return m.Gauge
}

func (m *MemStorage) SetGauge(n string, v float64) {
	m.Mutex.Lock()
	m.Gauge[n] = v
	m.Mutex.Unlock()
}

func (m *MemStorage) AddCounter(n string, v int64) {
	m.Mutex.Lock()
	val, ok := m.Counter[n]
	if ok {
		m.Counter[n] = val + v
	} else {
		m.Counter[n] = v
	}
	m.Mutex.Unlock()
}
