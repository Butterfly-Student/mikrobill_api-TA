package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	"prabogo/utils/activity"
)

type mikrotikPPPSecretAdapter struct {
	domain domain.Domain
}

func NewMikrotikPPPSecretAdapter(domain domain.Domain) inbound_port.MikrotikPPPSecretPort {
	return &mikrotikPPPSecretAdapter{
		domain: domain,
	}
}

// --- Secrets ---

func (h *mikrotikPPPSecretAdapter) MikrotikCreateSecret(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_create_secret")

	var input model.PPPSecretInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "Invalid inputs: " + err.Error()})
		return nil
	}

	res, err := h.domain.MikrotikPPPSecret().MikrotikCreateSecret(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: res})
	return nil
}

func (h *mikrotikPPPSecretAdapter) MikrotikGetSecret(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_get_secret")
	id := c.Param("id")

	res, err := h.domain.MikrotikPPPSecret().MikrotikGetSecret(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}

func (h *mikrotikPPPSecretAdapter) MikrotikUpdateSecret(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_update_secret")
	id := c.Param("id")

	var input model.PPPSecretUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	_, err := h.domain.MikrotikPPPSecret().MikrotikUpdateSecret(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Updated successfully"})
	return nil
}

func (h *mikrotikPPPSecretAdapter) MikrotikDeleteSecret(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_delete_secret")
	id := c.Param("id")

	err := h.domain.MikrotikPPPSecret().MikrotikDeleteSecret(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Deleted successfully"})
	return nil
}

func (h *mikrotikPPPSecretAdapter) MikrotikListSecrets(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_list_secrets")

	res, err := h.domain.MikrotikPPPSecret().MikrotikListSecrets(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}

