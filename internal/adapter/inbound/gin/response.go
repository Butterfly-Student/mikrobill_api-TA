package gin_inbound_adapter

import (
	"github.com/gin-gonic/gin"

	"prabogo/internal/model"
)

// SendResponse sends a structured JSON response
func SendResponse(c *gin.Context, statusCode int, data any, metadata *model.Metadata) {
	c.JSON(statusCode, model.Response{
		Success:  statusCode >= 200 && statusCode < 300,
		Data:     data,
		Metadata: metadata,
	})
}

// SendError sends a structured JSON error response
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, model.Response{
		Success: false,
		Error:   message,
	})
}

// SendAbort sends a structured JSON error response and aborts the request
func SendAbort(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, model.Response{
		Success: false,
		Error:   message,
	})
}
