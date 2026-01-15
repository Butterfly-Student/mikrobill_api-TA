package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitMikrotik, downInitMikrotik)
}

func upInitMikrotik(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS mikrotik (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			host INET NOT NULL,
			port INTEGER NOT NULL DEFAULT 8728,
			api_username TEXT NOT NULL,
			api_encrypted_password TEXT,
			keepalive BOOLEAN DEFAULT true,
			timeout INTEGER DEFAULT 300000,
			location VARCHAR(100),
			description TEXT,
			is_active BOOLEAN NOT NULL DEFAULT false,
			status mikrotik_status NOT NULL DEFAULT 'offline',
			last_sync TIMESTAMPTZ,
			
			-- Audit
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ,
			
			UNIQUE (tenant_id, host, port)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_mikrotik_tenant ON mikrotik(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_mikrotik_status ON mikrotik(status);`,
		`CREATE INDEX IF NOT EXISTS idx_mikrotik_is_active ON mikrotik(is_active);`,
		`CREATE INDEX IF NOT EXISTS idx_mikrotik_deleted_at ON mikrotik(deleted_at);`,

		`DROP TRIGGER IF EXISTS set_updated_at_mikrotik ON mikrotik;`,
		`CREATE TRIGGER set_updated_at_mikrotik
			BEFORE UPDATE ON mikrotik
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downInitMikrotik(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS mikrotik;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
