package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
)

type tenantAdapter struct {
	domainRegistry domain.Domain
}

func NewTenantAdapter(domainRegistry domain.Domain) inbound_port.TenantPort {
	return &tenantAdapter{
		domainRegistry: domainRegistry,
	}
}

func (a *tenantAdapter) CreateTenant(c *gin.Context) {
	var input model.TenantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	createdBy := userID.(uuid.UUID)

	tenant, err := a.domainRegistry.Tenant().CreateTenant(c.Request.Context(), input, createdBy)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to create tenant")
		return
	}

	SendResponse(c, http.StatusCreated, tenant, nil)
}

func (a *tenantAdapter) GetTenant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "Invalid tenant ID format",
		})
		return
	}

	tenant, err := a.domainRegistry.Tenant().GetTenant(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   "Tenant not found",
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    tenant,
	})
}

func (a *tenantAdapter) ListTenants(c *gin.Context) {
	var filter model.TenantFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	tenants, err := a.domainRegistry.Tenant().ListTenants(c.Request.Context(), filter)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to list tenants")
		return
	}

	SendResponse(c, http.StatusOK, tenants, &model.Metadata{
		Total:  int64(len(tenants)),
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
}

func (a *tenantAdapter) UpdateTenant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid ID",
			"details": "Invalid tenant ID format",
		})
		return
	}

	var input model.TenantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	updatedBy := userID.(uuid.UUID)

	tenant, err := a.domainRegistry.Tenant().UpdateTenant(c.Request.Context(), id, input, updatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   "Failed to update tenant",
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    tenant,
	})
}

func (a *tenantAdapter) DeleteTenant(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid ID",
			"details": "Invalid tenant ID format",
		})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	deletedBy := userID.(uuid.UUID)

	if err := a.domainRegistry.Tenant().DeleteTenant(c.Request.Context(), id, deletedBy); err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to delete tenant")
		return
	}

	SendResponse(c, http.StatusOK, nil, nil)
}

func (a *tenantAdapter) GetTenantStats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid ID",
			"details": "Invalid tenant ID format",
		})
		return
	}

	stats, err := a.domainRegistry.Tenant().GetTenantStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   "Failed to get tenant stats",
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    stats,
	})
}
