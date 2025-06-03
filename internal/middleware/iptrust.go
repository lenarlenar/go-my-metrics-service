package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

func TrustedSubnet(trustedCIDR string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if trustedCIDR == "" {
			c.Next()
			return
		}

		realIP := c.GetHeader("X-Real-IP")
		if !isIPTrusted(realIP, trustedCIDR) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

func isIPTrusted(realIP string, cidr string) bool {
	if cidr == "" {
		return true
	}

	ip := net.ParseIP(realIP)
	if ip == nil {
		return false
	}

	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	return subnet.Contains(ip)
}
