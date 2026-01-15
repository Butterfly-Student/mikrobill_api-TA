package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitLogging, downInitLogging)
}

func upInitLogging(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			
			action VARCHAR(100) NOT NULL, -- e.g., 'create', 'update', 'delete', 'login'
			resource_type VARCHAR(100) NOT NULL, -- e.g., 'customer', 'mikrotik', 'profile'
			resource_id UUID,
			
			description TEXT,
			old_values JSONB,
			new_values JSONB,
			
			ip_address INET,
			user_agent TEXT,
			
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_activity_logs_tenant ON activity_logs(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_user ON activity_logs(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_resource ON activity_logs(resource_type, resource_id);`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downInitLogging(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS activity_logs;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
