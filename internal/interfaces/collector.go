package interfaces

import "time"

type Collector interface {
	StartCollectAndUpdate(pollInterval int) *time.Ticker
}
