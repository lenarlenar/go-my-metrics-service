package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		log.I().Infoln(
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"status", c.Writer.Status(),
			"duration", duration,
			"size", c.Writer.Size(),
		)
	}
}

type GzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *GzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func GzipCompression() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			gw := gzip.NewWriter(c.Writer)
			defer gw.Close()
			c.Writer = &GzipWriter{c.Writer, gw}
			c.Header("Content-Encoding", "gzip")
		}
		c.Next()
	}
}

type GzipReader struct {
	io.ReadCloser
	reader *gzip.Reader
}

func (g *GzipReader) Read(p []byte) (int, error) {
	return g.reader.Read(p)
}

func GzipUnpack() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer gz.Close()

			c.Request.Body = &GzipReader{c.Request.Body, gz}
		}
		c.Next()
	}
}
