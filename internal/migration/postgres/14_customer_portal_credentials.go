package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCustomerPortalCredentials, downCustomerPortalCredentials)
}

func upCustomerPortalCredentials(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// 1. Add Portal Login Credentials (for customer portal access)
		`ALTER TABLE customers 
			ADD COLUMN IF NOT EXISTS portal_email VARCHAR(255),
			ADD COLUMN IF NOT EXISTS portal_password_hash TEXT;`,

		// Create unique index for portal_email per tenant (including deleted_at for soft delete support)
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_portal_email_unique 
			ON customers(tenant_id, portal_email, deleted_at) 
			WHERE portal_email IS NOT NULL;`,

		// Add index for fast portal login lookup
		`CREATE INDEX IF NOT EXISTS idx_customers_portal_email 
			ON customers(portal_email) 
			WHERE portal_email IS NOT NULL;`,

		// 2. Rename existing username to service_username for clarity
		`ALTER TABLE customers 
			RENAME COLUMN username TO service_username;`,

		// 3. Add Service Credentials (for PPPoE/Hotspot)
		`ALTER TABLE customers 
			ADD COLUMN IF NOT EXISTS service_password_encrypted TEXT,
			ADD COLUMN IF NOT EXISTS service_password_visible BOOLEAN DEFAULT false;`,

		// 4. Add Provisioning Status Fields
		`DO $$ 
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'provisioning_status') THEN
				CREATE TYPE provisioning_status AS ENUM ('pending', 'provisioning', 'active', 'failed');
			END IF;
		END $$;`,

		`ALTER TABLE customers 
			ADD COLUMN IF NOT EXISTS provisioning_status VARCHAR(20) DEFAULT 'pending',
			ADD COLUMN IF NOT EXISTS provisioning_error TEXT,
			ADD COLUMN IF NOT EXISTS provisioned_at TIMESTAMPTZ;`,

		// 5. Create Customer Sessions Table (for portal authentication)
		`CREATE TABLE IF NOT EXISTS customer_sessions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			tenant_id UUID NOT NULL,
			
			token_hash TEXT NOT NULL,
			refresh_token_hash TEXT,
			
			ip_address INET,
			user_agent TEXT,
			
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			
			UNIQUE(token_hash)
		);`,

		// Indexes for customer_sessions
		`CREATE INDEX IF NOT EXISTS idx_customer_sessions_customer 
			ON customer_sessions(customer_id);`,

		`CREATE INDEX IF NOT EXISTS idx_customer_sessions_tenant 
			ON customer_sessions(tenant_id);`,

		`CREATE INDEX IF NOT EXISTS idx_customer_sessions_token 
			ON customer_sessions(token_hash);`,

		`CREATE INDEX IF NOT EXISTS idx_customer_sessions_expires 
			ON customer_sessions(expires_at);`,

		// Update the unique constraint to use service_username instead of username
		`ALTER TABLE customers 
			DROP CONSTRAINT IF EXISTS customers_tenant_id_mikrotik_id_username_deleted_at_key;`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_service_username_unique 
			ON customers(tenant_id, mikrotik_id, service_username, deleted_at);`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downCustomerPortalCredentials(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// Drop customer_sessions table
		`DROP TABLE IF EXISTS customer_sessions;`,

		// Remove portal credentials
		`DROP INDEX IF EXISTS idx_customers_portal_email;`,
		`DROP INDEX IF EXISTS idx_customers_portal_email_unique;`,

		`ALTER TABLE customers 
			DROP COLUMN IF EXISTS portal_email,
			DROP COLUMN IF EXISTS portal_password_hash;`,

		// Rename service_username back to username
		`ALTER TABLE customers 
			RENAME COLUMN service_username TO username;`,

		// Remove service credential fields
		`ALTER TABLE customers 
			DROP COLUMN IF EXISTS service_password_encrypted,
			DROP COLUMN IF EXISTS service_password_visible;`,

		// Remove provisioning status fields
		`ALTER TABLE customers 
			DROP COLUMN IF EXISTS provisioning_status,
			DROP COLUMN IF EXISTS provisioning_error,
			DROP COLUMN IF EXISTS provisioned_at;`,

		`DROP TYPE IF EXISTS provisioning_status;`,

		// Restore old unique constraint
		`DROP INDEX IF EXISTS idx_customers_service_username_unique;`,

		`ALTER TABLE customers 
			ADD CONSTRAINT customers_tenant_id_mikrotik_id_username_deleted_at_key 
			UNIQUE (tenant_id, mikrotik_id, username, deleted_at);`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
