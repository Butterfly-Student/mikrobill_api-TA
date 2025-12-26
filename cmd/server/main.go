package main

import (
	"context"
	"fmt"
	"mikrobill/config"
	"mikrobill/internal/delivery/http/router"
	database "mikrobill/internal/infrastructure/db/postgres"
	pkg_logger "mikrobill/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	if err := pkg_logger.InitLogger(cfg.Logger.Environment); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer pkg_logger.Sync()

	pkg_logger.Info("Starting Mikrobill API Server")

	// Initialize database
	db, err := database.InitDatabase(cfg.Database)
	if err != nil {
		pkg_logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Initialize router
	r, err := router.NewRouter(db, cfg)
	if err != nil {
		pkg_logger.Fatal("Failed to initialize router", zap.Error(err))
	}
	engine := r.GetEngine()

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		pkg_logger.Info("Server starting",
			zap.String("host", cfg.Server.Host),
			zap.Int("port", cfg.Server.Port),
			zap.String("swagger", fmt.Sprintf("http://%s:%d/swagger/index.html", cfg.Server.Host, cfg.Server.Port)),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pkg_logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	pkg_logger.Info("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		pkg_logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	pkg_logger.Info("Server stopped gracefully")
}
