package gin_inbound_adapter

import (
	"context"

	"github.com/gin-gonic/gin"

	inbound_port "prabogo/internal/port/inbound"
)

func InitRoute(
	ctx context.Context,
	engine *gin.Engine,
	port inbound_port.HttpPort,
) {
	middlewareAdapter := port.Middleware()

	auth := engine.Group("/auth")
	auth.POST("/login", func(c *gin.Context) {
		port.Auth().Login(c)
	})
	auth.POST("/register", func(c *gin.Context) {
		port.Auth().Register(c)
	})

	internal := engine.Group("/internal")
	internal.Use(func(c *gin.Context) {
		if err := middlewareAdapter.InternalAuth(c); err != nil {
			return
		}
	})
	internal.POST("/client-upsert", func(c *gin.Context) {
		port.Client().Upsert(c)
	})
	internal.POST("/client-find", func(c *gin.Context) {
		port.Client().Find(c)
	})
	internal.DELETE("/client-delete", func(c *gin.Context) {
		port.Client().Delete(c)
	})

	// MikroTik routes
	internal.POST("/mikrotik", func(c *gin.Context) {
		port.Mikrotik().Create(c)
	})
	internal.POST("/mikrotik/list", func(c *gin.Context) {
		port.Mikrotik().List(c)
	})
	internal.GET("/mikrotik/active", func(c *gin.Context) {
		port.Mikrotik().GetActiveMikrotik(c)
	})
	internal.GET("/mikrotik/:id", func(c *gin.Context) {
		port.Mikrotik().GetByID(c)
	})
	internal.PUT("/mikrotik/:id", func(c *gin.Context) {
		port.Mikrotik().Update(c)
	})
	internal.DELETE("/mikrotik/:id", func(c *gin.Context) {
		port.Mikrotik().Delete(c)
	})
	internal.PATCH("/mikrotik/:id/status", func(c *gin.Context) {
		port.Mikrotik().UpdateStatus(c)
	})
	internal.PATCH("/mikrotik/:id/activate", func(c *gin.Context) {
		port.Mikrotik().SetActive(c)
	})

	// PPP Routes
	internal.POST("/ppp/secret", func(c *gin.Context) {
		port.PPP().CreateSecret(c)
	})
	internal.GET("/ppp/secret/:id", func(c *gin.Context) {
		port.PPP().GetSecret(c)
	})
	internal.PUT("/ppp/secret/:id", func(c *gin.Context) {
		port.PPP().UpdateSecret(c)
	})
	internal.DELETE("/ppp/secret/:id", func(c *gin.Context) {
		port.PPP().DeleteSecret(c)
	})
	internal.GET("/ppp/secret/list", func(c *gin.Context) {
		port.PPP().ListSecrets(c)
	})

	internal.POST("/ppp/profile", func(c *gin.Context) {
		port.PPP().CreateProfile(c)
	})
	internal.GET("/ppp/profile/:id", func(c *gin.Context) {
		port.PPP().GetProfile(c)
	})
	internal.PUT("/ppp/profile/:id", func(c *gin.Context) {
		port.PPP().UpdateProfile(c)
	})
	internal.DELETE("/ppp/profile/:id", func(c *gin.Context) {
		port.PPP().DeleteProfile(c)
	})
	internal.GET("/ppp/profile/list", func(c *gin.Context) {
		port.PPP().ListProfiles(c)
	})

	// Monitor Route (WebSocket)
	// Monitor route requires traffic interface name, usually protected or internal.
	// Placing in /api/v1/monitor or internal?
	// User request said "monitor-traffic interface" -> likely needs auth.
	// Internal seems safe for now or V1.
	// Monitor traffic is usually real-time, maybe V1?
	// But current requirement "routesnyaa.. websocket with gin".
	// Let's put it in internal for consistency with management, or add a new monitor group.
	// Given it's billing API, maybe /internal/monitor/traffic/:interface
	internal.GET("/monitor/traffic/:interface", func(c *gin.Context) {
		port.Monitor().StreamTraffic(c)
	})

	// Profile Routes (One-way sync DB -> MikroTik)
	internal.POST("/profile", func(c *gin.Context) {
		port.Profile().CreateProfile(c)
	})
	internal.GET("/profile/:id", func(c *gin.Context) {
		port.Profile().GetProfile(c)
	})
	internal.GET("/profile/list", func(c *gin.Context) {
		port.Profile().ListProfiles(c)
	})
	internal.PUT("/profile/:id", func(c *gin.Context) {
		port.Profile().UpdateProfile(c)
	})
	internal.DELETE("/profile/:id", func(c *gin.Context) {
		port.Profile().DeleteProfile(c)
	})

	// Customer Routes (One-way sync DB -> MikroTik)
	internal.POST("/customer", func(c *gin.Context) {
		port.Customer().CreateCustomer(c)
	})
	internal.GET("/customer/:id", func(c *gin.Context) {
		port.Customer().GetCustomer(c)
	})
	internal.GET("/customer/list", func(c *gin.Context) {
		port.Customer().ListCustomers(c)
	})
	internal.PUT("/customer/:id", func(c *gin.Context) {
		port.Customer().UpdateCustomer(c)
	})
	internal.DELETE("/customer/:id", func(c *gin.Context) {
		port.Customer().DeleteCustomer(c)
	})

	callbacks := engine.Group("/callbacks")
	// Callbacks might need auth or be whitelisted IPs. For now open or use middleware if needed.
	// User request showed no auth middleware on callback handler logic but typically it's open for Mikrotik to push.
	callbacks.POST("/pppoe/up", func(c *gin.Context) {
		port.Callback().HandlePPPoEUp(c)
	})
	callbacks.POST("/pppoe/down", func(c *gin.Context) {
		port.Callback().HandlePPPoEDown(c)
	})

	client := engine.Group("/v1")
	client.Use(func(c *gin.Context) {
		middlewareAdapter.ClientAuth(c)
	})
	client.GET("/ping", func(c *gin.Context) {
		port.Ping().GetResource(c)
	})
}
