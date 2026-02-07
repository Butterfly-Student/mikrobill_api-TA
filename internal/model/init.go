package model

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"MikrOps/utils/logger"
)

//go:generate mockgen -source=init.go -destination=../../tests/mocks/model/mock_init.go

// Environment variables
var (
	ENV_FILE   string
	APP_MODE   string
	DB_HOST    string
	DB_PORT    string
	DB_USER    string
	DB_PASSWORD string
	DB_NAME    string
	DB_SSLMODE string
	DB_MIGRATION_DIR string
	JWT_SECRET string
	INTERNAL_KEY string
	SERVER_PORT string
	OUTBOUND_DATABASE_DRIVER string
	OUTBOUND_MESSAGE_DRIVER  string
	OUTBOUND_CACHE_DRIVER    string
	INBOUND_HTTP_DRIVER      string
	INBOUND_MESSAGE_DRIVER   string
	MESSAGE_HOST     string
	MESSAGE_PORT     string
	MESSAGE_USER     string
	MESSAGE_PASSWORD string
	MESSAGE_VHOST    string
	CACHE_HOST     string
	CACHE_PORT     string
	CACHE_PASSWORD string
	SERVICE_CREDENTIAL_KEY          string
	UPSERT_CLIENT_MESSAGE_SUBSCRIBE string
	CORS_ALLOWED_ORIGINS            string
	ENCRYPTION_KEY                  string
)

func init() {
	LoadEnv()
	if APP_MODE == "production" {
		logger.GetLogger().Info("Application is running in production mode")
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func LoadEnv() {
	logger.GetLogger().Info("Loading environment variables")
	ENV_FILE = ".env"

	godotenv.Load(ENV_FILE)

	APP_MODE = os.Getenv("APP_MODE")
	DB_HOST = os.Getenv("DATABASE_HOST")
	DB_PORT = os.Getenv("DATABASE_PORT")
	DB_USER = os.Getenv("DATABASE_USERNAME")
	DB_PASSWORD = os.Getenv("DATABASE_PASSWORD")
	DB_NAME = getEnvOrDefault("DATABASE_NAME", "mikrotik_billing")
	DB_SSLMODE = getEnvOrDefault("DATABASE_SSLMODE", "disable")
	DB_MIGRATION_DIR = getEnvOrDefault("DATABASE_MIGRATION_DIR", "./internal/migration")

	JWT_SECRET = os.Getenv("JWT_SECRET")
	INTERNAL_KEY = getEnvOrDefault("INTERNAL_KEY", "")
	if INTERNAL_KEY == "" {
		logger.GetLogger().Warn("INTERNAL_KEY is not set, using default (not recommended)")
		INTERNAL_KEY = "default_internal_key"
	}
	if len(INTERNAL_KEY) < 32 {
		logger.GetLogger().Warn("INTERNAL_KEY too weak, must be at least 32 characters")
	}
	SERVER_PORT = getEnvOrDefault("SERVER_PORT", "8000")

	OUTBOUND_DATABASE_DRIVER = getEnvOrDefault("OUTBOUND_DATABASE_DRIVER", "postgres")
	OUTBOUND_MESSAGE_DRIVER = getEnvOrDefault("OUTBOUND_MESSAGE_DRIVER", "rabbitmq")
	OUTBOUND_CACHE_DRIVER = getEnvOrDefault("OUTBOUND_CACHE_DRIVER", "redis")
	INBOUND_HTTP_DRIVER = getEnvOrDefault("INBOUND_HTTP_DRIVER", "gin")
	INBOUND_MESSAGE_DRIVER = getEnvOrDefault("INBOUND_MESSAGE_DRIVER", "rabbitmq")

	MESSAGE_HOST = os.Getenv("MESSAGE_HOST")
	MESSAGE_PORT = getEnvOrDefault("MESSAGE_PORT", "5672")
	MESSAGE_USER = os.Getenv("MESSAGE_USER")
	MESSAGE_PASSWORD = os.Getenv("MESSAGE_PASSWORD")
	MESSAGE_VHOST = getEnvOrDefault("MESSAGE_VHOST", "")

	CACHE_HOST = getEnvOrDefault("CACHE_HOST", "redis")
	CACHE_PORT = getEnvOrDefault("CACHE_PORT", "6379")
	CACHE_PASSWORD = getEnvOrDefault("CACHE_PASSWORD", "")

	SERVICE_CREDENTIAL_KEY = os.Getenv("SERVICE_CREDENTIAL_KEY")
	if SERVICE_CREDENTIAL_KEY == "" {
		logger.GetLogger().Warn("SERVICE_CREDENTIAL_KEY not set, using default")
		SERVICE_CREDENTIAL_KEY = "default_key_please_change_in_production"
	}
	if len(SERVICE_CREDENTIAL_KEY) != 32 {
		logger.GetLogger().Warn("SERVICE_CREDENTIAL_KEY must be exactly 32 bytes for AES-256")
	}
	UPSERT_CLIENT_MESSAGE_SUBSCRIBE = getEnvOrDefault("UPSERT_CLIENT_MESSAGE_SUBSCRIBE", "client.upsert.subscribe")
	CORS_ALLOWED_ORIGINS = getEnvOrDefault("CORS_ALLOWED_ORIGINS", "*")
	if CORS_ALLOWED_ORIGINS == "*" {
		logger.GetLogger().Warn("CORS_ALLOWED_ORIGINS is set to *, not recommended for production")
	}
	ENCRYPTION_KEY = getEnvOrDefault("ENCRYPTION_KEY", "default_key_please_change_in_production")
	if len(ENCRYPTION_KEY) != 32 {
		logger.GetLogger().Warn("ENCRYPTION_KEY must be exactly 32 bytes for AES-256")
	}

	logger.GetLogger().Info("Environment variables loaded successfully")
}

func Validate() error {
	var errors []string

	if APP_MODE != "development" && APP_MODE != "release" && APP_MODE != "test" {
		errors = append(errors, "APP_MODE must be one of: development, release, test")
	}

	if APP_MODE == "production" {
		if JWT_SECRET == "" {
			errors = append(errors, "JWT_SECRET is required in production")
		} else if len(JWT_SECRET) < 32 {
			errors = append(errors, "JWT_SECRET must be at least 32 characters")
		}

		if INTERNAL_KEY == "" {
			errors = append(errors, "INTERNAL_KEY is required in production")
		} else if len(INTERNAL_KEY) < 32 {
			errors = append(errors, "INTERNAL_KEY must be at least 32 characters")
		}

		if SERVICE_CREDENTIAL_KEY == "" {
			errors = append(errors, "SERVICE_CREDENTIAL_KEY is required in production")
		} else if len(SERVICE_CREDENTIAL_KEY) != 32 {
			errors = append(errors, "SERVICE_CREDENTIAL_KEY must be exactly 32 bytes for AES-256")
		}

		if CORS_ALLOWED_ORIGINS == "*" {
			errors = append(errors, "CORS_ALLOWED_ORIGINS is set to *, not recommended for production")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %v", errors)
	}

	return nil
}

// AutoMigrate runs database migrations for GORM models only
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Tenant{},
		&TenantUser{},
		&User{},
		&Role{},
		&Mikrotik{},
		&MikrotikProfile{},
		&MikrotikProfilePPPoE{},
		&MikrotikProfileHotspot{},
		&MikrotikProfileStaticIP{},
		&Customer{},
		&CustomerService{},
		&ActivityLog{},
		&SuperadminLog{},
		&GlobalSetting{},
		&RateLimitRule{},
		&RateLimitViolation{},
		&IPAMPool{},
		&IPAllocation{},
	)
}
