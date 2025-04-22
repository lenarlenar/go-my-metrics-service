package log

import (
	"sync"

	"go.uber.org/zap"
)

// LocalLogger - обертка для zap.SugaredLogger, предоставляющая удобный интерфейс для логирования.
type LocalLogger struct {
	*zap.SugaredLogger
}

var (
	// instance - единственный экземпляр LocalLogger.
	instance *LocalLogger
	// once - используется для гарантированной инициализации logger-а только один раз.
	once sync.Once
)

// I - возвращает единственный экземпляр LocalLogger. Логгер инициализируется только один раз.
func I() *LocalLogger {
	once.Do(func() {
		// I - возвращает единственный экземпляр LocalLogger. Логгер инициализируется только один раз.
		zapLogger, err := zap.NewDevelopment()
		if err != nil {
			// Если не удается создать логгер, программа завершится с ошибкой.
			panic(err)
		}
		// Ожидание завершения всех операций с логгером перед его инициализацией.
		defer zapLogger.Sync()

		// Ожидание завершения всех операций с логгером перед его инициализацией.
		instance = &LocalLogger{zapLogger.Sugar()}

	})

	// Возвращаем единственный экземпляр логгера.
	return instance
}
