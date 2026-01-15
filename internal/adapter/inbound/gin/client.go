package gin_inbound_adapter

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	"prabogo/utils/activity"
)

type clientAdapter struct {
	domain domain.Domain
}

func NewClientAdapter(
	domain domain.Domain,
) inbound_port.ClientHttpPort {
	return &clientAdapter{
		domain: domain,
	}
}

func (h *clientAdapter) Upsert(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_client_upsert")
	var payload []model.ClientInput
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return nil
	}
	ctx = context.WithValue(ctx, activity.Payload, payload)

	results, err := h.domain.Client().Upsert(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    results,
	})
	return nil
}

func (h *clientAdapter) Find(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_client_find_by_filter")
	var payload model.ClientFilter
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return nil
	}
	ctx = context.WithValue(ctx, activity.Payload, payload)

	results, err := h.domain.Client().FindByFilter(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    results,
	})
	return nil
}

func (h *clientAdapter) Delete(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_client_delete_by_filter")
	var payload model.Request
	payload.Data = model.ClientFilter{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return nil
	}
	ctx = context.WithValue(ctx, activity.Payload, payload)

	err := h.domain.Client().DeleteByFilter(ctx, payload.Data.(model.ClientFilter))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
	})
	return nil
}

