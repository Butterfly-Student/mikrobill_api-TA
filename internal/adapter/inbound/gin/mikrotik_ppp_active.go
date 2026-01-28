package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/activity"
)

type mikrotikPPPActiveAdapter struct {
	domain domain.Domain
}

func NewMikrotikPPPActiveAdapter(domain domain.Domain) inbound_port.MikrotikPPPActivePort {
	return &mikrotikPPPActiveAdapter{
		domain: domain,
	}
}

func (h *mikrotikPPPActiveAdapter) MikrotikListActive(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_list_active")

	res, err := h.domain.MikrotikPPPActive().MikrotikListActive(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}

func (h *mikrotikPPPActiveAdapter) MikrotikGetActive(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_get_active")
	id := c.Param("id")

	res, err := h.domain.MikrotikPPPActive().MikrotikGetActive(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}
