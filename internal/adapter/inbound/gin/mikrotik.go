package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	"prabogo/utils/activity"
)

type mikrotikAdapter struct {
	domain domain.Domain
}

func NewMikrotikAdapter(
	domain domain.Domain,
) inbound_port.MikrotikHttpPort {
	return &mikrotikAdapter{
		domain: domain,
	}
}

func (h *mikrotikAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_create")

	var input model.MikrotikInput
	if err := c.ShouldBindJSON(&input); err != nil {
		SendError(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return nil
	}

	result, err := h.domain.Mikrotik().Create(ctx, input)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error())
		return nil
	}

	SendResponse(c, http.StatusCreated, result, nil)
	return nil
}

func (h *mikrotikAdapter) GetByID(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_get_by_id")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid ID format",
		})
		return nil
	}

	result, err := h.domain.Mikrotik().GetByID(ctx, id)
	if err != nil {
		SendError(c, http.StatusNotFound, err.Error())
		return nil
	}

	SendResponse(c, http.StatusOK, result, nil)
	return nil
}

func (h *mikrotikAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_list")

	var filter model.MikrotikFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		// If no body, list all
		filter = model.MikrotikFilter{}
	}

	results, err := h.domain.Mikrotik().List(ctx, filter)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error())
		return nil
	}

	SendResponse(c, http.StatusOK, results, &model.Metadata{
		Total: int64(len(results)),
	})
	return nil
}

func (h *mikrotikAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_update")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid ID format",
		})
		return nil
	}

	var input model.MikrotikUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return nil
	}

	result, err := h.domain.Mikrotik().Update(ctx, id, input)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error())
		return nil
	}

	SendResponse(c, http.StatusOK, result, nil)
	return nil
}

func (h *mikrotikAdapter) Delete(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_delete")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid ID format",
		})
		return nil
	}

	err = h.domain.Mikrotik().Delete(ctx, id)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error())
		return nil
	}

	SendResponse(c, http.StatusOK, nil, nil)
	return nil
}

func (h *mikrotikAdapter) UpdateStatus(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_update_status")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid ID format",
		})
		return nil
	}

	var input struct {
		Status model.MikrotikStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid request body",
		})
		return nil
	}

	err = h.domain.Mikrotik().UpdateStatus(ctx, id, input.Status)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error())
		return nil
	}

	SendResponse(c, http.StatusOK, nil, nil)
	return nil
}

func (h *mikrotikAdapter) GetActiveMikrotik(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_get_active")

	result, err := h.domain.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error())
		return nil
	}

	if result == nil {
		SendError(c, http.StatusNotFound, "No active mikrotik found")
		return nil
	}

	SendResponse(c, http.StatusOK, result, nil)
	return nil
}

func (h *mikrotikAdapter) SetActive(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_set_active")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid ID format",
		})
		return nil
	}

	err = h.domain.Mikrotik().SetActive(ctx, id)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err.Error())
		return nil
	}

	SendResponse(c, http.StatusOK, nil, nil)
	return nil
}

