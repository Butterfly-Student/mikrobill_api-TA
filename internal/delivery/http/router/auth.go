// File: internal/router/routes_auth.go
package router

import (
	"mikrobill/internal/delivery/http/handler"
	"mikrobill/internal/delivery/http/middleware"
	"mikrobill/internal/port/repository"
	"mikrobill/internal/port/service"
	"mikrobill/internal/usecase"
)

// setupAuthRoutes configures authentication routes
func (r *Router) setupAuthRoutes() {
	// Initialize services - langsung instantiate tanpa constructor
	passwordService := &service.PasswordService{}
	jwtService := service.NewJWTService(r.config.JWT.SecretKey, r.config.JWT.TokenDuration)

	// Initialize repositories
	userRepo := repository.NewUserRepository(r.db)
	roleRepo := repository.NewRoleRepository(r.db)

	// Initialize usecase
	authUsecase := usecase.NewAuthUsecase(
		userRepo,
		roleRepo,
		passwordService,
		jwtService,
	)

	// Initialize handler
	authHandler := handler.NewAuthHandler(authUsecase)

	// Setup route groups
	v1 := r.engine.Group("/api/v1")
	{
		// Public auth routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
		}

		// Protected routes (authentication required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService))
		{
			authGroup := protected.Group("/auth")
			{
				authGroup.GET("/profile", authHandler.GetProfile)
				authGroup.POST("/change-password", authHandler.ChangePassword)
				authGroup.POST("/logout", authHandler.Logout)
			}
		}
	}
}

