package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetLogger() *LocalLogger {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync()

	return &LocalLogger{zapLogger.Sugar()}
}

type LocalLogger struct {
	zapLogger *zap.SugaredLogger
}

func (l *LocalLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.zapLogger.Fatalw(msg, keysAndValues...)
}

func (l *LocalLogger) Infoln(args ...interface{}) {
	l.zapLogger.Infoln(args)
}

func (l *LocalLogger) GetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		l.zapLogger.Infoln(
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"status", c.Writer.Status(),
			"duration", duration,
			"size", c.Writer.Size(),
		)
	}
}