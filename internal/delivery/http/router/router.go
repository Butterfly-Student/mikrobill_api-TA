// File: internal/router/router.go
package router

import (
	"fmt"
	"mikrobill/config"
	"mikrobill/internal/delivery/http/middleware"
	"mikrobill/pkg/filelog"
	pkg_logger "mikrobill/pkg/logger"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

type Router struct {
	engine   *gin.Engine
	db       *gorm.DB
	config   *config.Config
	enforcer *casbin.Enforcer
}

// NewRouter creates a new router instance with all dependencies
func NewRouter(db *gorm.DB, cfg *config.Config) (*Router, error) {
	// Initialize Casbin adapter
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// Initialize Casbin enforcer
	enforcer, err := casbin.NewEnforcer("config/casbin_model.conf", adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// Load policies
	if err := enforcer.LoadPolicy(); err != nil {
		pkg_logger.Warn("Failed to load Casbin policies, starting with empty policy")
	}

	// Initialize file logger for API access logs
	if err := filelog.Init(); err != nil {
		pkg_logger.Warn("Failed to initialize file logger")
	}

	r := &Router{
		engine:   gin.New(),
		db:       db,
		config:   cfg,
		enforcer: enforcer,
	}

	// Setup middlewares
	r.setupGlobalMiddlewares()

	// Setup swagger and health check
	r.setupSwaggerAndHealth()

	// Register all routes
	r.setupRoutes()

	return r, nil
}

// setupGlobalMiddlewares configures global middlewares
func (r *Router) setupGlobalMiddlewares() {
	r.engine.Use(
		middleware.RecoveryMiddleware(),
		// middleware.LoggerMiddleware(),
		middleware.CORSMiddleware(),
		middleware.ErrorHandler(),
		middleware.RateLimitMiddleware(300),
	)
}

// setupSwaggerAndHealth configures Swagger documentation and health check endpoint
func (r *Router) setupSwaggerAndHealth() {
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// setupRoutes registers all application routes
func (r *Router) setupRoutes() {
	// Setup auth routes
	r.setupAuthRoutes()

	// Setup application routes
	r.setupAppRoutes()

	// Add more route groups here as your application grows
	// r.setupUserRoutes()
	// r.setupProductRoutes()
	// etc.
}

// Accessor methods
func (r *Router) DB() *gorm.DB                  { return r.db }
func (r *Router) Config() *config.Config        { return r.config }
func (r *Router) Enforcer() *casbin.Enforcer    { return r.enforcer }
func (r *Router) Engine() *gin.Engine           { return r.engine }
func (r *Router) GetEngine() *gin.Engine        { return r.engine }
func (r *Router) GetEnforcer() *casbin.Enforcer { return r.enforcer }

// Helper middleware methods
func (r *Router) ParseID() gin.HandlerFunc     { return r.parseID() }
func (r *Router) ParseUserID() gin.HandlerFunc { return r.parseUserID() }
