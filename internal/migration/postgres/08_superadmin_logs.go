package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSuperadminLogs, downSuperadminLogs)
}

func upSuperadminLogs(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS superadmin_logs (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			superadmin_id UUID REFERENCES users(id) ON DELETE SET NULL,
			
			action VARCHAR(100) NOT NULL, -- e.g., 'create_tenant', 'suspend_tenant', 'update_system_config'
			target_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
			resource_type VARCHAR(100), -- e.g., 'tenant', 'user', 'system_config'
			resource_id UUID,
			
			description TEXT,
			old_values JSONB,
			new_values JSONB,
			
			ip_address INET,
			user_agent TEXT,
			
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_superadmin_logs_admin ON superadmin_logs(superadmin_id);`,
		`CREATE INDEX IF NOT EXISTS idx_superadmin_logs_tenant ON superadmin_logs(target_tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_superadmin_logs_created_at ON superadmin_logs(created_at);`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downSuperadminLogs(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS superadmin_logs;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
