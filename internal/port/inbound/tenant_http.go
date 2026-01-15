package inbound_port

import "github.com/gin-gonic/gin"

// TenantPort defines HTTP handlers for tenant endpoints
type TenantPort interface {
	CreateTenant(c *gin.Context)
	GetTenant(c *gin.Context)
	ListTenants(c *gin.Context)
	UpdateTenant(c *gin.Context)
	DeleteTenant(c *gin.Context)
	GetTenantStats(c *gin.Context)
}
