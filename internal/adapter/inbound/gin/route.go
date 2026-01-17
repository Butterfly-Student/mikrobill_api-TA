package gin_inbound_adapter

import (
	"context"

	"github.com/gin-gonic/gin"

	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
)

func InitRoute(
	ctx context.Context,
	engine *gin.Engine,
	port inbound_port.HttpPort,
) {
	middlewareAdapter := port.Middleware()

	// 1. Recovery (using Gin's default)
	engine.Use(gin.Recovery())

	// 2. Request ID
	engine.Use(middlewareAdapter.RequestID())

	// 3. Logging
	engine.Use(middlewareAdapter.ZapLogger())

	// 4. CORS
	engine.Use(middlewareAdapter.CORS())

	// 5. Rate Limiting (Global)
	// engine.Use(middlewareAdapter.RateLimit())

	// 9. Request Validation (Global Validator)
	engine.Use(middlewareAdapter.Validator())

	// Versioning Group
	v1 := engine.Group("/v1")

	// Auth group
	auth := v1.Group("/auth")
	auth.POST("/login", func(c *gin.Context) {
		port.Auth().Login(c)
	})
	auth.POST("/register", func(c *gin.Context) {
		port.Auth().Register(c)
	})
	auth.POST("/logout", func(c *gin.Context) {
		port.Auth().Logout(c)
	})
	auth.POST("/refresh", func(c *gin.Context) {
		port.Auth().RefreshToken(c)
	})
	auth.GET("/profile", func(c *gin.Context) {
		port.Auth().GetProfile(c)
	})

	// Internal group (Admin/Management)
	internal := v1.Group("/internal")
	internal.Use(func(c *gin.Context) {
		// 6. Authentication (Internal & User support)
		if err := middlewareAdapter.InternalAuth(c); err != nil {
			return
		}
	})

	// 12. Tenant Management (GLOBAL - Restricted to SuperAdmin)
	tenant := internal.Group("/tenant")
	tenant.Use(middlewareAdapter.RequireRole(string(model.UserRoleSuperAdmin)))
	{
		tenant.POST("", func(c *gin.Context) { port.Tenant().CreateTenant(c) })
		tenant.GET("/list", func(c *gin.Context) { port.Tenant().ListTenants(c) })
		tenant.GET("/:id", func(c *gin.Context) { port.Tenant().GetTenant(c) })
		tenant.PUT("/:id", func(c *gin.Context) { port.Tenant().UpdateTenant(c) })
		tenant.DELETE("/:id", func(c *gin.Context) { port.Tenant().DeleteTenant(c) })
		tenant.GET("/:id/stats", func(c *gin.Context) { port.Tenant().GetTenantStats(c) })
	}

	// User Management
	users := internal.Group("/users")
	{
		users.POST("", func(c *gin.Context) { port.User().CreateUser(c) })
		users.GET("/list", func(c *gin.Context) { port.User().ListUsers(c) })
		users.GET("/:id", func(c *gin.Context) { port.User().GetUser(c) })
		users.PUT("/:id", func(c *gin.Context) { port.User().UpdateUser(c) })
		users.DELETE("/:id", func(c *gin.Context) { port.User().DeleteUser(c) })
		users.POST("/:id/assign-role", func(c *gin.Context) { port.User().AssignRole(c) })
		users.POST("/:id/assign-tenant", func(c *gin.Context) { port.User().AssignToTenant(c) })
	}

	// Isolated Resources (Need Tenant Context)
	resources := internal.Group("/")
	resources.Use(middlewareAdapter.TenantContext())
	resources.Use(middlewareAdapter.RequireTenantAccess())

	// Resource routes
	resources.POST("/client-upsert", func(c *gin.Context) {
		port.Client().Upsert(c)
	})
	resources.POST("/client-find", func(c *gin.Context) {
		port.Client().Find(c)
	})
	resources.DELETE("/client-delete", func(c *gin.Context) {
		port.Client().Delete(c)
	})

	// MikroTik routes
	mikrotik := resources.Group("/mikrotik")
	{
		mikrotik.POST("", func(c *gin.Context) { port.Mikrotik().Create(c) })
		mikrotik.POST("/list", func(c *gin.Context) { port.Mikrotik().List(c) })
		mikrotik.GET("/active", func(c *gin.Context) { port.Mikrotik().GetActiveMikrotik(c) })
		mikrotik.GET("/:id", func(c *gin.Context) { port.Mikrotik().GetByID(c) })
		mikrotik.PUT("/:id", func(c *gin.Context) { port.Mikrotik().Update(c) })
		mikrotik.DELETE("/:id", func(c *gin.Context) { port.Mikrotik().Delete(c) })
		mikrotik.PATCH("/:id/status", func(c *gin.Context) { port.Mikrotik().UpdateStatus(c) })
		mikrotik.PATCH("/:id/activate", func(c *gin.Context) { port.Mikrotik().SetActive(c) })
	}

	// PPP Routes
	ppp := resources.Group("/ppp")
	{
		ppp.POST("/secret", func(c *gin.Context) { port.MikrotikPPPSecret().MikrotikCreateSecret(c) })
		ppp.GET("/secret/:id", func(c *gin.Context) { port.MikrotikPPPSecret().MikrotikGetSecret(c) })
		ppp.PUT("/secret/:id", func(c *gin.Context) { port.MikrotikPPPSecret().MikrotikUpdateSecret(c) })
		ppp.DELETE("/secret/:id", func(c *gin.Context) { port.MikrotikPPPSecret().MikrotikDeleteSecret(c) })
		ppp.GET("/secret/list", func(c *gin.Context) { port.MikrotikPPPSecret().MikrotikListSecrets(c) })

		ppp.POST("/profile", func(c *gin.Context) { port.MikrotikPPPProfile().MikrotikCreateProfile(c) })
		ppp.GET("/profile/:id", func(c *gin.Context) { port.MikrotikPPPProfile().MikrotikGetProfile(c) })
		ppp.PUT("/profile/:id", func(c *gin.Context) { port.MikrotikPPPProfile().MikrotikUpdateProfile(c) })
		ppp.DELETE("/profile/:id", func(c *gin.Context) { port.MikrotikPPPProfile().MikrotikDeleteProfile(c) })
		ppp.GET("/profile/list", func(c *gin.Context) { port.MikrotikPPPProfile().MikrotikListProfiles(c) })
	}

	// Profile & Customer Routes
	resources.POST("/profile", func(c *gin.Context) { port.Profile().CreateProfile(c) })
	resources.GET("/profile/:id", func(c *gin.Context) { port.Profile().GetProfile(c) })
	resources.GET("/profile/list", func(c *gin.Context) { port.Profile().ListProfiles(c) })
	resources.PUT("/profile/:id", func(c *gin.Context) { port.Profile().UpdateProfile(c) })
	resources.DELETE("/profile/:id", func(c *gin.Context) { port.Profile().DeleteProfile(c) })

	customer := resources.Group("/customer")
	{
		customer.POST("", func(c *gin.Context) { port.Customer().CreateCustomer(c) })
		customer.GET("/:id", func(c *gin.Context) { port.Customer().GetCustomer(c) })
		customer.GET("/list", func(c *gin.Context) { port.Customer().ListCustomers(c) })
		customer.PUT("/:id", func(c *gin.Context) { port.Customer().UpdateCustomer(c) })
		customer.DELETE("/:id", func(c *gin.Context) { port.Customer().DeleteCustomer(c) })
		customer.GET("/:id/traffic/stream", func(c *gin.Context) { port.Monitor().StreamTraffic(c) })
		customer.GET("/:id/ping", func(c *gin.Context) { port.Monitor().PingCustomer(c) })
		customer.GET("/:id/ping/stream", func(c *gin.Context) { port.Monitor().StreamPing(c) })
	}

	// Monitor
	resources.GET("/monitor/traffic/:interface", func(c *gin.Context) {
		port.Monitor().StreamTraffic(c)
	})

	// Callbacks (Outside v1 potentially, but let's keep consistency for now)
	callbacks := v1.Group("/callbacks")
	callbacks.POST("/pppoe/up", func(c *gin.Context) { port.Callback().HandlePPPoEUp(c) })
	callbacks.POST("/pppoe/down", func(c *gin.Context) { port.Callback().HandlePPPoEDown(c) })

	// Legacy client compat if needed, but versioned is better
	clientCompat := v1.Group("/client")
	clientCompat.Use(func(c *gin.Context) {
		middlewareAdapter.ClientAuth(c)
	})
	clientCompat.GET("/ping", func(c *gin.Context) {
		port.Ping().GetResource(c)
	})
}
