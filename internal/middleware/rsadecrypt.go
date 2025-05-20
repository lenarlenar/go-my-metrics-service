package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RSADecrypt(privKey *rsa.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		if privKey == nil {
			c.Next()
			return
		}

		if c.Request.Header.Get("Content-Type") != "application/octet-stream" {
			c.Next()
			return
		}

		encryptedData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
			return
		}

		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, encryptedData)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "RSA decryption failed"})
			return
		}

		// Заменяем тело запроса на расшифрованные данные
		c.Request.Body = io.NopCloser(bytes.NewReader(decrypted))
		c.Request.ContentLength = int64(len(decrypted))
		c.Request.Header.Set("Content-Type", "application/json") // важно для дальнейшей обработки

		c.Next()
	}
}
