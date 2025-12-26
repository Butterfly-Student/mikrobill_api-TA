// internal/router/helpers.go
package router

import (
	"mikrobill/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (r *Router) parseID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			utils.ErrorResponse(c, 400, "Invalid ID parameter", err)
			c.Abort()
			return
		}
		c.Set("id", id)
		c.Next()
	}
}

func (r *Router) parseUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("user_id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			utils.ErrorResponse(c, 400, "Invalid user ID parameter", err)
			c.Abort()
			return
		}
		c.Set("user_id", id)
		c.Next()
	}
}