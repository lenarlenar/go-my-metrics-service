package interfaces

import "github.com/gin-gonic/gin"

type Service interface {
	IndexHandler(c *gin.Context)
	ValueHandler(c *gin.Context)
	UpdateHandler(c *gin.Context)
	ValueJSONHandler(c *gin.Context)
	UpdateJSONHandler(c *gin.Context)
}
