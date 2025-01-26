package log

import (
	"sync"

	"go.uber.org/zap"
)

type LocalLogger struct {
	*zap.SugaredLogger
}

var (
	instance *LocalLogger
	once     sync.Once
)

func I() *LocalLogger {
	once.Do(func() {

		zapLogger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		defer zapLogger.Sync()

		instance = &LocalLogger{zapLogger.Sugar()}

	})
	return instance
}
