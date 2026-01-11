package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateCustomersTable, downCreateCustomersTable)
}

func upCreateCustomersTable(ctx context.Context, tx *sql.Tx) error {
	// 1. Ensure ENUM types exist
	queries := []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'customer_status') THEN CREATE TYPE customer_status AS ENUM ('active', 'suspended', 'inactive', 'pending'); END IF; END $$;`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'service_type') THEN CREATE TYPE service_type AS ENUM ('pppoe', 'hotspot', 'static_ip'); END IF; END $$;`,
	}
	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}

	// 2. Create customers table
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS customers (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
			
			-- Basic info
			username VARCHAR(100) NOT NULL,
			name VARCHAR(255) NOT NULL,
			phone VARCHAR(20) NOT NULL,
			email VARCHAR(255),
			address TEXT,
			mikrotik_object_id VARCHAR(50) NOT NULL,
			
			-- Service configuration
			service_type service_type NOT NULL,
			
			-- Network info
			assigned_ip INET,
			mac_address MACADDR,
			interface VARCHAR(20),
			last_online TIMESTAMPTZ,
			last_ip INET,
			
			-- Status & billing
			status customer_status DEFAULT 'inactive',
			auto_suspension BOOLEAN DEFAULT true,
			billing_day INTEGER DEFAULT 15,
			join_date TIMESTAMPTZ DEFAULT now(),
			
			customer_notes TEXT,
			
			created_at TIMESTAMPTZ DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now(),
			
			UNIQUE (mikrotik_id, username),
			UNIQUE (mikrotik_id, phone)
		);
	`)
	if err != nil {
		return err
	}

	// 3. Create Indices
	indexQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_customers_mikrotik ON customers(mikrotik_id);`,
		`CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);`,
		`CREATE INDEX IF NOT EXISTS idx_customers_service_type ON customers(service_type);`,
		`CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);`,
		`CREATE INDEX IF NOT EXISTS idx_customers_mac ON customers(mac_address);`,
	}
	for _, query := range indexQueries {
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}

	// 4. Create Trigger
	_, err = tx.Exec(`
		DROP TRIGGER IF EXISTS set_updated_at_customers ON customers;
		CREATE TRIGGER set_updated_at_customers
			BEFORE UPDATE ON customers
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();
	`)
	if err != nil {
		return err
	}

	return nil
}

func downCreateCustomersTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS customers;`)
	if err != nil {
		return err
	}
	// We might want to keep the types if they are used elsewhere, or drop them.
	// For now, assuming they might be shared or re-created, we'll leave them or drop them if this was the only user.
	// User didn't specify down migration specifics, so standard drop table is safe.
	return nil
}
