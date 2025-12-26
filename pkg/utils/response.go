// internal/utils/response.go
package utils

import (

	"github.com/gin-gonic/gin"
)


type APIResponse struct {
    Success bool        `json:"success" example:"true"`
    Message string      `json:"message" example:"Operation successful"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}



func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Error:   errMsg,
	})
}