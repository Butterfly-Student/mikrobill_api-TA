package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	"prabogo/utils/activity"
)

/*
type- [x] Domain Unit Testing <!-- id: 18 -->
  - [x] Auth Domain Tests
  - [x] Customer Domain Tests
  - [x] Mikrotik Domain Tests
  - [x] Monitor Domain Tests
  - [x] PPP Domain Tests (Secret & Profile)
  - [x] Profile Domain Tests
  - [x] Verify mock generation for all ports
  - [x] Final verification of all tests
*/
type mikrotikPPPProfileAdapter struct {
	domain domain.Domain
}

func NewMikrotikPPPProfileAdapter(domain domain.Domain) inbound_port.MikrotikPPPProfilePort {
	return &mikrotikPPPProfileAdapter{
		domain: domain,
	}
}

// --- Profiles ---

func (h *mikrotikPPPProfileAdapter) MikrotikCreateProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_create_profile")

	var input model.PPPProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	res, err := h.domain.MikrotikPPPProfile().MikrotikCreateProfile(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: res})
	return nil
}

func (h *mikrotikPPPProfileAdapter) MikrotikGetProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_get_profile")
	id := c.Param("id")

	res, err := h.domain.MikrotikPPPProfile().MikrotikGetProfile(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}

func (h *mikrotikPPPProfileAdapter) MikrotikUpdateProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_update_profile")
	id := c.Param("id")

	var input model.PPPProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	_, err := h.domain.MikrotikPPPProfile().MikrotikUpdateProfile(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Updated successfully"})
	return nil
}

func (h *mikrotikPPPProfileAdapter) MikrotikDeleteProfile(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_delete_profile")
	id := c.Param("id")

	err := h.domain.MikrotikPPPProfile().MikrotikDeleteProfile(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Deleted successfully"})
	return nil
}

func (h *mikrotikPPPProfileAdapter) MikrotikListProfiles(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_list_profiles")

	res, err := h.domain.MikrotikPPPProfile().MikrotikListProfiles(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}

