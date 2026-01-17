package model

import (
	"MikrOps/utils/logger"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AutoMigrate - Run all migrations
func AutoMigrate(db *gorm.DB) error {
	models := []interface{}{
		&Tenant{},
		&Role{},
		&User{},
		&TenantUser{},
		&UserSession{},
		&SessionActivity{},
		&Mikrotik{},
		&MikrotikProfile{},
		&MikrotikProfilePPPoE{},
		&MikrotikProfileHotspot{},
		&MikrotikProfileStaticIP{},
		&Customer{},
		&CustomerService{},
		&ActivityLog{},
		&SuperadminLog{},
		&RateLimitRule{},
		&RateLimitViolation{},
		&GlobalSetting{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			logger.Error("failed to migrate model",
				zap.String("model", fmt.Sprintf("%T", model)),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}
