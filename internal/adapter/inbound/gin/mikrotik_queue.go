package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/activity"
)

type mikrotikQueueAdapter struct {
	domain domain.Domain
}

func NewMikrotikQueueAdapter(domain domain.Domain) inbound_port.MikrotikQueuePort {
	return &mikrotikQueueAdapter{
		domain: domain,
	}
}

func (h *mikrotikQueueAdapter) MikrotikCreateQueue(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_create_queue")

	var input model.QueueSimpleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: "Invalid inputs: " + err.Error()})
		return nil
	}

	res, err := h.domain.MikrotikQueue().MikrotikCreateQueue(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: res})
	return nil
}

func (h *mikrotikQueueAdapter) MikrotikGetQueue(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_get_queue")
	id := c.Param("id")

	res, err := h.domain.MikrotikQueue().MikrotikGetQueue(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}

func (h *mikrotikQueueAdapter) MikrotikUpdateQueue(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_update_queue")
	id := c.Param("id")

	var input model.QueueSimpleUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	_, err := h.domain.MikrotikQueue().MikrotikUpdateQueue(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Updated successfully"})
	return nil
}

func (h *mikrotikQueueAdapter) MikrotikDeleteQueue(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_delete_queue")
	id := c.Param("id")

	err := h.domain.MikrotikQueue().MikrotikDeleteQueue(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: "Deleted successfully"})
	return nil
}

func (h *mikrotikQueueAdapter) MikrotikListQueues(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_mikrotik_list_queues")

	res, err := h.domain.MikrotikQueue().MikrotikListQueues(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	c.JSON(http.StatusOK, model.Response{Success: true, Data: res})
	return nil
}
