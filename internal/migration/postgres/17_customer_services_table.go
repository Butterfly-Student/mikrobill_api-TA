package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateCustomerServicesTable, downCreateCustomerServicesTable)
}

func upCreateCustomerServicesTable(ctx context.Context, tx *sql.Tx) error {
	// Create customer_services table (the most important one)
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS customer_services (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			profile_id UUID NOT NULL REFERENCES mikrotik_profiles(id) ON DELETE RESTRICT,
			price NUMERIC(12,2) NOT NULL,
			tax_rate NUMERIC(5,2) DEFAULT 11.00,
			start_date DATE NOT NULL,
			end_date DATE,
			status service_status NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (customer_id, profile_id)
		);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downCreateCustomerServicesTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS customer_services;`)
	if err != nil {
		return err
	}
	return nil
}
