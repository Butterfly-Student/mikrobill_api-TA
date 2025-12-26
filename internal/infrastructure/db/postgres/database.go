// config/database.go
package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"mikrobill/config"
	pkg_logger "mikrobill/pkg/logger"

	"gorm.io/gorm/logger"

	"go.uber.org/zap"
)

// Custom GORM Logger
type GormLogger struct {
	SlowThreshold time.Duration
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	pkg_logger.Info(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	pkg_logger.Warn(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	pkg_logger.Error(fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil && err != gorm.ErrRecordNotFound {
		pkg_logger.Error("GORM Error",
			zap.Error(err),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	} else if elapsed > l.SlowThreshold && l.SlowThreshold != 0 {
		pkg_logger.Warn("Slow SQL Query",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	} else {
		pkg_logger.Debug("SQL Query",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	}
}

func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: &GormLogger{
			SlowThreshold: 200 * time.Millisecond,
		},
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pkg_logger.Info("Database connected successfully",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Database),
	)

	return db, nil
}

// Auto migrate tables
// func AutoMigrate(db *gorm.DB) error {
// 	logger.Info("Running auto migration...")

// 	// Import domain models
// 	err := db.AutoMigrate(
// 		&domain.User{},
// 		&domain.Role{},
// 		&domain.Permission{},
// 		&domain.Resource{},
// 		&domain.Action{},
// 		&domain.Mikrotik{},
// 		&domain.CasbinRule{},
// 	)

// 	if err != nil {
// 		return fmt.Errorf("failed to auto migrate: %w", err)
// 	}

// 	logger.Info("Auto migration completed successfully")
// 	return nil
// }
