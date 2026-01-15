package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
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

	var input model.CustomerInput
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

	var input model.CustomerInput
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
