package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitCustomers, downInitCustomers)
}

func upInitCustomers(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// 1. Customers Table
		`CREATE TABLE IF NOT EXISTS customers (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
			
			-- Basic info
			username VARCHAR(100) NOT NULL,
			name VARCHAR(255) NOT NULL,
			phone VARCHAR(20) NOT NULL,
			email VARCHAR(255),
			address TEXT,
			mikrotik_object_id VARCHAR(50), -- ID from MikroTik
			
			-- Service configuration
			service_type service_type NOT NULL,
			
			-- Network info
			assigned_ip INET,
			mac_address MACADDR,
			interface VARCHAR(50),
			last_online TIMESTAMPTZ,
			last_ip INET,
			
			-- Status & billing
			status customer_status DEFAULT 'inactive',
			auto_suspension BOOLEAN DEFAULT true,
			billing_day INTEGER DEFAULT 1,
			join_date DATE DEFAULT CURRENT_DATE,
			
			customer_notes TEXT,
			
			-- Audit
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ,
			
			UNIQUE (tenant_id, mikrotik_id, username, deleted_at),
			UNIQUE (tenant_id, phone, deleted_at)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_customers_tenant ON customers(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_customers_mikrotik ON customers(mikrotik_id);`,
		`CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);`,
		`CREATE INDEX IF NOT EXISTS idx_customers_deleted_at ON customers(deleted_at);`,

		`DROP TRIGGER IF EXISTS set_updated_at_customers ON customers;`,
		`CREATE TRIGGER set_updated_at_customers
			BEFORE UPDATE ON customers
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,

		// 2. Customer Services Table
		`CREATE TABLE IF NOT EXISTS customer_services (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			profile_id UUID NOT NULL REFERENCES mikrotik_profiles(id) ON DELETE RESTRICT,
			
			price NUMERIC(15,2) NOT NULL,
			tax_rate NUMERIC(5,2) DEFAULT 0.00,
			
			start_date DATE NOT NULL,
			end_date DATE,
			status service_status NOT NULL DEFAULT 'active',
			
			-- Audit
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_cust_services_tenant ON customer_services(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_cust_services_customer ON customer_services(customer_id);`,
		`CREATE INDEX IF NOT EXISTS idx_cust_services_profile ON customer_services(profile_id);`,

		`DROP TRIGGER IF EXISTS set_updated_at_customer_services ON customer_services;`,
		`CREATE TRIGGER set_updated_at_customer_services
			BEFORE UPDATE ON customer_services
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downInitCustomers(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS customer_services;`,
		`DROP TABLE IF EXISTS customers;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
