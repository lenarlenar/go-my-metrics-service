package workerpool

import (
	"sync"

	"github.com/lenarlenar/go-my-metrics-service/internal/agent/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
	"github.com/lenarlenar/go-my-metrics-service/internal/sender"
)

type Pool struct {
	wg    sync.WaitGroup
	jobs  chan map[string]model.Metrics
	flags flags.Flags
}

// New создает пул с заданным числом воркеров и буфером.
func New(flags flags.Flags, workerCount int) *Pool {
	p := &Pool{
		jobs:  make(chan map[string]model.Metrics, workerCount),
		flags: flags,
	}

	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	return p
}

// Submit добавляет задачу в очередь. false, если очередь переполнена.
func (p *Pool) Submit(metrics map[string]model.Metrics) bool {
	select {
	case p.jobs <- metrics:
		return true
	default:
		return false
	}
}

// Shutdown завершает все воркеры.
func (p *Pool) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for data := range p.jobs {
		sender.Send(p.flags, data)
	}
}
