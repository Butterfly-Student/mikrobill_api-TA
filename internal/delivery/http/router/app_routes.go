package router

import (
	"context"
	"log"
	"mikrobill/internal/delivery/http/handler"
	"mikrobill/internal/port/repository"
	"mikrobill/internal/usecase"
	"mikrobill/pkg/pub_sub"
)

// setupAppRoutes configures application routes (Customers, Profiles, Monitoring)
func (r *Router) setupAppRoutes() {
	// 1. Initialize Infrastructure (Redis)
	redisPublisher := pub_sub.NewRedisPublisher(&r.config.Redis)

	// 2. Initialize Repositories
	customerRepo := repository.NewDatabaseCustomerRepository(r.db)
	mikrotikRepo := repository.NewMikrotikRepository(r.db)
	profileRepo := repository.NewDatabaseProfileRepository(r.db)

	// 3. Initialize Services (Usecases)
	// Mikrotik UseCase (to get client)
	mikrotikUseCase := usecase.NewMikrotikUseCase(mikrotikRepo, r.config.Crypto.EncryptionKey)

	// Get Active Mikrotik Client
	mtClient, err := mikrotikUseCase.GetMikrotikClient()
	if err != nil {
		log.Printf("[Router] WARNING: Failed to get active Mikrotik client: %v. Feature dealing with Mikrotik will be disabled until restart/reload.", err)
	}

	customerService := usecase.NewCustomerService(customerRepo, profileRepo, mtClient)
	profileService := usecase.NewProfileService(profileRepo, mikrotikUseCase)
	trafficService := usecase.NewOnDemandTrafficService(mtClient, customerRepo, redisPublisher)

	// 4. Initialize Handlers
	wsHandler := handler.NewWebSocketHandler()
	// Run generic broadcaster
	go wsHandler.Broadcaster()

	// Start Redis Subscriber for Realtime Events
	go func() {
		bgCtx := context.Background()

		pubsub := redisPublisher.GetClient().Subscribe(bgCtx, "mikrotik:events")
		defer pubsub.Close()

		ch := pubsub.Channel()

		log.Println("[Router] Started Redis Subscriber for 'mikrotik:events'")

		for msg := range ch {
			// Broadcast message to all connected WebSocket clients
			// log.Printf("[Router] Received Redis message: %s", msg.Payload)
			wsHandler.GetBroadcastChannel() <- []byte(msg.Payload)
		}
	}()

	callbackHandler := handler.NewCallbackHandler(customerRepo, redisPublisher)
	customerHandler := handler.NewCustomerHandler(customerService)
	profileHandler := handler.NewProfileHandler(profileService)
	trafficHandler := handler.NewTrafficMonitorHandler(trafficService, customerRepo, mtClient)
	mikrotikHandler := handler.NewMikrotikHandler(mikrotikUseCase)

	// 5. Register Routes based on user request

	// WebSocket endpoint
	r.engine.GET("/ws", wsHandler.HandleWS)

	// API routes
	api := r.engine.Group("/api")
	{
		// Common Routes
		api.GET("/mikrotiks", mikrotikHandler.ListMikrotiks)
		api.POST("/mikrotiks", mikrotikHandler.CreateMikrotik)

		// Callback routes (MikroTik WebHooks)
		callbacks := api.Group("/callbacks")
		{
			callbacks.POST("/pppoe-up", callbackHandler.HandlePPPoEUp)
			callbacks.POST("/pppoe-down", callbackHandler.HandlePPPoEDown)
		}

		// Customer routes (CRUD)
		customers := api.Group("/customers")
		{
			// CRUD operations (handled by CustomerHandler)
			customers.GET("", customerHandler.ListCustomers)
			customers.POST("", customerHandler.CreateCustomer)
			customers.GET("/:id", customerHandler.GetCustomer)
			customers.PUT("/:id", customerHandler.UpdateCustomer)
			customers.DELETE("/:id", customerHandler.DeleteCustomer)

			// Monitoring Specifics (handled by TrafficMonitorHandler)
			// These extend the customer resource
			customers.GET("/:id/ping", trafficHandler.GetPingHandler().PingCustomerByID)
			customers.GET("/:id/ping/ws", trafficHandler.GetPingHandler().PingCustomerStream)
			customers.GET("/:id/traffic/ws", trafficHandler.StreamCustomerTraffic)
		}

		// Monitor routes
		monitor := api.Group("/monitor")
		{
			monitor.GET("/status", trafficHandler.GetStatus)
		}

		// Reload customers route - keeping if logic exists, though deprecated in OnDemand
		api.POST("/reload-customers", trafficHandler.ReloadCustomers)

		// Profile routes (CRUD and Sync)
		profiles := api.Group("/profiles")
		{
			profiles.GET("", profileHandler.ListProfiles)
			profiles.POST("", profileHandler.CreateProfile)
			profiles.GET("/:id", profileHandler.GetProfile)
			profiles.PUT("/:id", profileHandler.UpdateProfile)
			profiles.DELETE("/:id", profileHandler.DeleteProfile)
			profiles.POST("/:id/sync", profileHandler.SyncProfile)
			profiles.POST("/sync-all/:mikrotik_id", profileHandler.SyncAllProfiles)
		}
	}

	// Protected routes example (if needed, reuse middleware)
	// protected := api.Group("")
	// protected.Use(middleware.AuthMiddleware(jwtService))
	// Note: Authentication middleware is setup in auth routes separately.
	// If the user wants these routes protected, we should ask or apply it.
	// The user snippet didn't explicitly show auth middleware application for these groups,
	// but mostly mentioned global middleware.
	// For now, I will implement as requested (publicly accessible or relying on global).
	// However, usually API routes are protected.
	// The user request has: router.Use(middleware.CORS()), router.Use(gin.Recovery()) which are global.

	log.Println("[Router] App routes registered")
}
