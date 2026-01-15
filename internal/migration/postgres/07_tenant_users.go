package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upTenantUsers, downTenantUsers)
}

func upTenantUsers(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tenant_users (
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
			is_primary BOOLEAN DEFAULT false,
			is_active BOOLEAN DEFAULT true,
			
			-- Audit
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			
			PRIMARY KEY (tenant_id, user_id)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_tenant_users_user ON tenant_users(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant ON tenant_users(tenant_id);`,

		`DROP TRIGGER IF EXISTS set_updated_at_tenant_users ON tenant_users;`,
		`CREATE TRIGGER set_updated_at_tenant_users
			BEFORE UPDATE ON tenant_users
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downTenantUsers(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS tenant_users;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
