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
	// Create customers table
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS customers (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
			username VARCHAR(100) NOT NULL,
			full_name VARCHAR(150) NOT NULL,
			phone VARCHAR(20),
			email VARCHAR(100),
			address VARCHAR(255),
			status customer_status NOT NULL DEFAULT 'inactive',
			join_date TIMESTAMPTZ NOT NULL DEFAULT now(),
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (mikrotik_id, username),
			UNIQUE (mikrotik_id, phone)
		);
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
	return nil
}
