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
TenantContext() gin.HandlerFunc
RequireTenantAccess() gin.HandlerFunc
RequireRole(roles ...string) gin.HandlerFunc
Validator() gin.HandlerFunc
CustomerAuth() gin.HandlerFunc

PublicRegistrationRateLimit() gin.HandlerFunc
CustomerLoginRateLimit() gin.HandlerFunc
CustomerPortalRateLimit() gin.HandlerFunc

SecurityHeaders() gin.HandlerFunc
OriginValidation() gin.HandlerFunc
}

