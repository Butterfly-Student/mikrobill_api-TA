package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/activity"
)

type mikrotikPPPInactiveAdapter struct {
	domain domain.Domain
}

func NewMikrotikPPPInactiveAdapter(domain domain.Domain) inbound_port.MikrotikPPPInactivePort {
	return &mikrotikPPPInactiveAdapter{
		domain: domain,
	}
}

func (h *mikrotikPPPInactiveAdapter) MikrotikListInactive(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_list_inactive")

	res, err := h.domain.MikrotikPPPInactive().MikrotikListInactive(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}
