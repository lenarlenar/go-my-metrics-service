package middleware

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
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

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

func GzipCompression() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			gw := gzipWriterPool.Get().(*gzip.Writer)
			gw.Reset(c.Writer)
			defer func() {
				gw.Close()
				gzipWriterPool.Put(gw)
			}()
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

func calculateHash(data, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
func CheckHash(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {

		if secretKey == "" {
			log.I().Warn("secretKey не задан")
			c.Next()
			return
		}

		hash := c.GetHeader("HashSHA256")
		if hash == "" {
			log.I().Warn("HashSHA256 не задан в headers")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.I().Warnf("не удалось прочитать Body: %v/n", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Request.Body = io.NopCloser(strings.NewReader(string(data)))

		expectedHash := calculateHash(data, []byte(secretKey))
		if hash != expectedHash {
			log.I().Warn("HashSHA256: хеши не совпали")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		c.Next()

		responseData := []byte(c.Writer.Header().Get("Content-Type") + c.Request.URL.Path + c.Request.URL.RawQuery)
		responseHash := calculateHash(responseData, []byte(secretKey))
		c.Writer.Header().Set("HashSHA256", responseHash)
	}
}
