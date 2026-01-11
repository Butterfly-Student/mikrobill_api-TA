package inbound_port

import "github.com/gin-gonic/gin"

type CallbackHttpPort interface {
	HandlePPPoEUp(c *gin.Context)
	HandlePPPoEDown(c *gin.Context)
}
