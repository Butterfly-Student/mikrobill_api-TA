package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
)

type customerAdapter struct {
	domainRegistry domain.Domain
}

func NewCustomerAdapter(domainRegistry domain.Domain) inbound_port.CustomerPort {
	return &customerAdapter{
		domainRegistry: domainRegistry,
	}
}

func (a *customerAdapter) CreateCustomer(ctx any) error {
	c := ctx.(*gin.Context)

	var input model.CreateCustomerRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to bind request")
	}

	customer, err := a.domainRegistry.Customer().CreateCustomer(c, input)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to create customer")
		return stacktrace.Propagate(err, "failed to create customer")
	}

	SendResponse(c, http.StatusCreated, customer, nil)

	return nil
}

func (a *customerAdapter) GetCustomer(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	customer, err := a.domainRegistry.Customer().GetCustomer(c, id)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to get customer")
		return stacktrace.Propagate(err, "failed to get customer")
	}

	SendResponse(c, http.StatusOK, customer, nil)

	return nil
}

func (a *customerAdapter) ListCustomers(ctx any) error {
	c := ctx.(*gin.Context)

	customers, err := a.domainRegistry.Customer().ListCustomers(c)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to list customers")
		return stacktrace.Propagate(err, "failed to list customers")
	}

	SendResponse(c, http.StatusOK, customers, &model.Metadata{
		Total: int64(len(customers)),
	})

	return nil
}

func (a *customerAdapter) UpdateCustomer(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	var input model.CreateCustomerRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to bind request")
	}

	customer, err := a.domainRegistry.Customer().UpdateCustomer(c, id, input)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to update customer")
		return stacktrace.Propagate(err, "failed to update customer")
	}

	SendResponse(c, http.StatusOK, customer, nil)

	return nil
}

func (a *customerAdapter) DeleteCustomer(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	err := a.domainRegistry.Customer().DeleteCustomer(c, id)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to delete customer")
		return stacktrace.Propagate(err, "failed to delete customer")
	}

	SendResponse(c, http.StatusOK, nil, nil)

	return nil
}

// PublicRegister handles public registration without authentication
func (a *customerAdapter) PublicRegister(ctx any) error {
	c := ctx.(*gin.Context)

	// Get tenant slug from URL parameter
	tenantSlug := c.Param("tenant_slug")
	if tenantSlug == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "tenant slug is required",
		})
		return nil
	}

	// Find tenant by slug
	tenant, err := a.domainRegistry.Tenant().GetBySlug(c.Request.Context(), tenantSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   "tenant not found",
		})
		return stacktrace.Propagate(err, "tenant not found")
	}

	// Check if tenant is active
	if !tenant.IsActive || tenant.Status != "active" {
		c.JSON(http.StatusForbidden, model.Response{
			Success: false,
			Error:   "tenant is not accepting registrations",
		})
		return nil
	}

	// Parse request body
	var req model.PublicRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return stacktrace.Propagate(err, "failed to bind request")
	}

	// Call domain logic
	prospect, err := a.domainRegistry.Customer().RegisterProspect(c.Request.Context(), tenant.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return stacktrace.Propagate(err, "failed to register prospect")
	}

	// Return success response
	c.JSON(http.StatusCreated, model.Response{
		Success: true,
		Data: gin.H{
			"message":  "Registration successful. Please wait for admin approval.",
			"prospect": prospect,
		},
	})

	return nil
}

// ListProspects lists all prospects (admin only)
func (a *customerAdapter) ListProspects(ctx any) error {
	c := ctx.(*gin.Context)

	prospects, err := a.domainRegistry.Customer().ListProspects(c.Request.Context())
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to list prospects")
		return stacktrace.Propagate(err, "failed to list prospects")
	}

	SendResponse(c, http.StatusOK, gin.H{
		"prospects": prospects,
		"total":     len(prospects),
	}, nil)

	return nil
}

// ApproveProspect approves a prospect and provisions to MikroTik (admin only)
func (a *customerAdapter) ApproveProspect(ctx any) error {
	c := ctx.(*gin.Context)

	var req model.ApproveProspectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return stacktrace.Propagate(err, "failed to bind request")
	}

	customer, err := a.domainRegistry.Customer().ApproveProspect(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return stacktrace.Propagate(err, "failed to approve prospect")
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data: gin.H{
			"message":  "Prospect approved and provisioned to MikroTik",
			"customer": customer,
		},
	})

	return nil
}

// RejectProspect rejects a prospect (admin only)
func (a *customerAdapter) RejectProspect(ctx any) error {
	c := ctx.(*gin.Context)
	customerID := c.Param("id")

	var req struct {
		Reason *string `json:"reason,omitempty"`
	}
	_ = c.ShouldBindJSON(&req)

	err := a.domainRegistry.Customer().RejectProspect(c.Request.Context(), customerID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return stacktrace.Propagate(err, "failed to reject prospect")
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data: gin.H{
			"message": "Prospect rejected successfully",
		},
	})

	return nil
}
