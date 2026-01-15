package inbound_port

import "github.com/gin-gonic/gin"

type MiddlewareHttpPort interface {
	InternalAuth(a any) error
	ClientAuth(a any) error
	UserAuth(a any) error

	RequestID() gin.HandlerFunc
	ZapLogger() gin.HandlerFunc
	CORS() gin.HandlerFunc
	RateLimit() gin.HandlerFunc
	TenantContext() gin.HandlerFunc // Automatic tenant resolution from context
	RequireTenantAccess() gin.HandlerFunc
	RequireRole(roles ...string) gin.HandlerFunc
	Validator() gin.HandlerFunc
}
