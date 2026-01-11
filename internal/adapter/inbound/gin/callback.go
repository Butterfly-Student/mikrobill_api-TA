package gin_inbound_adapter

import (
	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"

	"github.com/gin-gonic/gin"
)

type callbackHandler struct {
	domain domain.Domain
}

func NewCallbackAdapter(domain domain.Domain) inbound_port.CallbackHttpPort {
	return &callbackHandler{
		domain: domain,
	}
}

func (h *callbackHandler) HandlePPPoEUp(c *gin.Context) {
	var input model.PPPoEUpInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	err := h.domain.Customer().HandlePPPoEUp(c, input)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}

func (h *callbackHandler) HandlePPPoEDown(c *gin.Context) {
	var input model.PPPoEDownInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"status": "error", "message": err.Error()})
		return
	}

	err := h.domain.Customer().HandlePPPoEDown(c, input)
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "success"})
}
