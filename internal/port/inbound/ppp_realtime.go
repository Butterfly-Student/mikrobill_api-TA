package inbound_port

import "github.com/gin-gonic/gin"

type PPPRealtimePort interface {
	StreamPPPActive(c *gin.Context) error
	StreamPPPInactive(c *gin.Context) error
}
