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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create customer",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to create customer")
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "customer created successfully",
		"data":    customer,
	})

	return nil
}

func (a *customerAdapter) GetCustomer(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	customer, err := a.domainRegistry.Customer().GetCustomer(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get customer",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to get customer")
	}

	c.JSON(http.StatusOK, gin.H{
		"data": customer,
	})

	return nil
}

func (a *customerAdapter) ListCustomers(ctx any) error {
	c := ctx.(*gin.Context)

	customers, err := a.domainRegistry.Customer().ListCustomers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list customers",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to list customers")
	}

	c.JSON(http.StatusOK, gin.H{
		"data": customers,
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update customer",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to update customer")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "customer updated successfully",
		"data":    customer,
	})

	return nil
}

func (a *customerAdapter) DeleteCustomer(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	err := a.domainRegistry.Customer().DeleteCustomer(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete customer",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to delete customer")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "customer deleted successfully",
	})

	return nil
}
