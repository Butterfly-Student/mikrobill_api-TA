package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"

	"MikrOps/internal/model"
	seeder_pkg "MikrOps/internal/seeder"
	"MikrOps/utils"
	"MikrOps/utils/logger"

	"go.uber.org/zap"
)

func InitDatabase(ctx context.Context, outboundDatabaseDriver string) *gorm.DB {
	var dialector gorm.Dialector

	switch outboundDatabaseDriver {
	case "postgres":
		dialector = postgres.Open(utils.GetDatabaseString())
	default:
		logger.Fatal("unsupported database driver", zap.String("driver", outboundDatabaseDriver))
	}

	// Configure GORM
	config := &gorm.Config{
		Logger: gorm_logger.Default.LogMode(getLogLevel()),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(dialector, config)
	if err != nil {
		logger.Fatal("failed to open database", zap.Error(err))
	}

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("failed to get sql.DB", zap.Error(err))
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}

	// Reset Database (Optional: Enable via config or just do it as requested for now)
	// WARN: This will wipe all data! Use with caution.
	if err := ResetDatabase(ctx, db); err != nil {
		logger.Fatal("failed to reset database", zap.Error(err))
	}

	// Run migrations
	if err := model.AutoMigrate(db); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	// Run seeders
	seeder := seeder_pkg.NewSeeder(db)
	if err := seeder.Seed(); err != nil {
		logger.Fatal("failed to seed data", zap.Error(err))
	}

	logger.Info("database initialized successfully")
	return db
}

func getLogLevel() gorm_logger.LogLevel {
	if os.Getenv("APP_MODE") == "release" {
		return gorm_logger.Error
	}
	return gorm_logger.Info
}

// ResetDatabase drops the public schema and recreates it to ensure a clean state
// ONLY RUN THIS IN DEVELOPMENT/TESTING ENVIRONMENTS
func ResetDatabase(ctx context.Context, db *gorm.DB) error {
	if os.Getenv("DATABASE_FORCE_RESET") != "true" {
		return nil
	}

	appMode := os.Getenv("APP_MODE")
	logger.Info("ResetDatabase called", zap.String("app_mode", appMode))

	// Safety check: Prevent running in production
	if appMode == "production" {
		logger.Warn("ResetDatabase called in production mode - skipping for safety")
		return nil
	}

	logger.Info("Resetting database schema...")

	// PostgreSQL specific command to drop and recreate public schema
	// Split into separate transactions for safety
	if err := db.Exec("DROP SCHEMA public CASCADE").Error; err != nil {
		return fmt.Errorf("failed to drop schema: %w", err)
	}

	if err := db.Exec("CREATE SCHEMA public").Error; err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	if err := db.Exec("SET search_path TO public").Error; err != nil {
		return fmt.Errorf("failed to set search path: %w", err)
	}

	// Re-enable UUID extension which was dropped with the schema
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" SCHEMA public").Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	// Create Enum Types (Required for Postgres + GORM enum mapping)
	if err := db.Exec("CREATE TYPE user_role AS ENUM ('superadmin', 'admin', 'technician', 'viewer')").Error; err != nil {
		return fmt.Errorf("failed to create user_role enum: %w", err)
	}

	if err := db.Exec("CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'banned')").Error; err != nil {
		return fmt.Errorf("failed to create user_status enum: %w", err)
	}

	if err := db.Exec("CREATE TYPE mikrotik_status AS ENUM ('online', 'offline', 'error')").Error; err != nil {
		return fmt.Errorf("failed to create mikrotik_status enum: %w", err)
	}

	if err := db.Exec("CREATE TYPE profile_type AS ENUM ('pppoe', 'hotspot', 'static_ip')").Error; err != nil {
		return fmt.Errorf("failed to create profile_type enum: %w", err)
	}

	if err := db.Exec("CREATE TYPE customer_status AS ENUM ('active', 'inactive', 'suspended', 'pending')").Error; err != nil {
		return fmt.Errorf("failed to create customer_status enum: %w", err)
	}

	if err := db.Exec("CREATE TYPE service_type AS ENUM ('pppoe', 'hotspot', 'static_ip')").Error; err != nil {
		return fmt.Errorf("failed to create service_type enum: %w", err)
	}

	if err := db.Exec("CREATE TYPE service_status AS ENUM ('active', 'suspended', 'terminated')").Error; err != nil {
		return fmt.Errorf("failed to create service_status enum: %w", err)
	}

	// Verify creation
	var count int64
	db.Raw("SELECT count(*) FROM pg_type WHERE typname = 'user_role'").Scan(&count)
	logger.Info("Verification: user_role type count", zap.Int64("count", count))

	db.Raw("SELECT count(*) FROM pg_extension WHERE extname = 'uuid-ossp'").Scan(&count)
	logger.Info("Verification: uuid-ossp extension count", zap.Int64("count", count))

	// Restore default grants (optional but good practice)
	if err := db.Exec("GRANT ALL ON SCHEMA public TO public").Error; err != nil {
		logger.Warn("failed to grant schema privileges", zap.Error(err))
		// Continue anyway
	}

	logger.Info("Database schema reset successfully")
	return nil
}
