package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitSystem, downInitSystem)
}

func upInitSystem(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// 1. Extensions
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,

		// 2. Common Functions
		`CREATE OR REPLACE FUNCTION set_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		// 3. Shared ENUM Types
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'mikrotik_status') THEN
				CREATE TYPE mikrotik_status AS ENUM ('offline', 'online', 'error');
			END IF;
		END $$;`,

		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'profile_type') THEN
				CREATE TYPE profile_type AS ENUM ('pppoe', 'hotspot', 'static_ip');
			END IF;
		END $$;`,

		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'customer_status') THEN
				CREATE TYPE customer_status AS ENUM ('active', 'suspended', 'inactive', 'pending');
			END IF;
		END $$;`,

		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'service_type') THEN
				CREATE TYPE service_type AS ENUM ('pppoe', 'hotspot', 'static_ip');
			END IF;
		END $$;`,

		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'service_status') THEN
				CREATE TYPE service_status AS ENUM ('active', 'suspended', 'terminated');
			END IF;
		END $$;`,

		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
				CREATE TYPE user_role AS ENUM ('superadmin', 'admin', 'technician', 'viewer');
			END IF;
		END $$;`,

		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
				CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'banned');
			END IF;
		END $$;`,

		// 4. Tenants Table
		`CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(200) NOT NULL,
			subdomain VARCHAR(100) UNIQUE,
			company_name VARCHAR(200),
			phone VARCHAR(50),
			address TEXT,
			timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
			is_active BOOLEAN DEFAULT TRUE,
			status VARCHAR(20) DEFAULT 'active',
			
			-- Limit Management
			max_mikrotiks INTEGER DEFAULT 3,
			max_network_users INTEGER DEFAULT 50,
			max_staff_users INTEGER DEFAULT 5,
			
			-- Features & Metadata
			features JSONB DEFAULT '{"api_access": true, "reports": true, "backup": true}',
			metadata JSONB DEFAULT '{}',
			
			-- Audit & Lifecycle
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			suspended_at TIMESTAMPTZ,
			deleted_at TIMESTAMPTZ
		);`,

		`CREATE INDEX IF NOT EXISTS idx_tenants_subdomain ON tenants(subdomain);`,
		`CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);`,
		`CREATE INDEX IF NOT EXISTS idx_tenants_deleted_at ON tenants(deleted_at);`,

		`DROP TRIGGER IF EXISTS set_updated_at_tenants ON tenants;`,
		`CREATE TRIGGER set_updated_at_tenants
			BEFORE UPDATE ON tenants
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downInitSystem(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS tenants;`,
		`DROP TYPE IF EXISTS user_status;`,
		`DROP TYPE IF EXISTS user_role;`,
		`DROP TYPE IF EXISTS service_status;`,
		`DROP TYPE IF EXISTS service_type;`,
		`DROP TYPE IF EXISTS customer_status;`,
		`DROP TYPE IF EXISTS profile_type;`,
		`DROP TYPE IF EXISTS mikrotik_status;`,
		`DROP FUNCTION IF EXISTS set_updated_at;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
