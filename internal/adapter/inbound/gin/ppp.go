package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	"prabogo/utils/activity"
)

type pppAdapter struct {
	domain domain.Domain
}

func NewPPPAdapter(domain domain.Domain) inbound_port.PPPPort {
	return &pppAdapter{
		domain: domain,
	}
}

// --- Secrets ---

func (h *pppAdapter) CreateSecret(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_create_secret")

	var input model.PPPSecretInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "Invalid inputs: " + err.Error()})
		return nil
	}

	res, err := h.domain.PPP().CreateSecret(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: res})
	return nil
}

func (h *pppAdapter) GetSecret(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_get_secret")
	id := c.Param("id")

	res, err := h.domain.PPP().GetSecret(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}

func (h *pppAdapter) UpdateSecret(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_update_secret")
	id := c.Param("id")

	var input model.PPPSecretUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	_, err := h.domain.PPP().UpdateSecret(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Updated successfully"})
	return nil
}

func (h *pppAdapter) DeleteSecret(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_delete_secret")
	id := c.Param("id")

	err := h.domain.PPP().DeleteSecret(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Deleted successfully"})
	return nil
}

func (h *pppAdapter) ListSecrets(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_list_secrets")

	res, err := h.domain.PPP().ListSecrets(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}

// --- Profiles ---

func (h *pppAdapter) CreateProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_create_profile")

	var input model.PPPProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	res, err := h.domain.PPP().CreateProfile(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: res})
	return nil
}

func (h *pppAdapter) GetProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_get_profile")
	id := c.Param("id")

	res, err := h.domain.PPP().GetProfile(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}

func (h *pppAdapter) UpdateProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_update_profile")
	id := c.Param("id")

	var input model.PPPProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	_, err := h.domain.PPP().UpdateProfile(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Updated successfully"})
	return nil
}

func (h *pppAdapter) DeleteProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_delete_profile")
	id := c.Param("id")

	err := h.domain.PPP().DeleteProfile(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Deleted successfully"})
	return nil
}

func (h *pppAdapter) ListProfiles(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_ppp_list_profiles")

	res, err := h.domain.PPP().ListProfiles(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}
