package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddMikrotikObjectId, downAddMikrotikObjectId)
}

func upAddMikrotikObjectId(ctx context.Context, tx *sql.Tx) error {
	// Add mikrotik_object_id to customer_services table
	_, err := tx.Exec(`
		ALTER TABLE customer_services
		ADD COLUMN IF NOT EXISTS mikrotik_object_id VARCHAR(50);
	`)
	if err != nil {
		return err
	}

	// Add index for better query performance
	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_customer_services_mikrotik_object_id 
		ON customer_services(mikrotik_object_id);
	`)
	if err != nil {
		return err
	}

	return nil
}

func downAddMikrotikObjectId(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP INDEX IF EXISTS idx_customer_services_mikrotik_object_id;
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		ALTER TABLE customer_services
		DROP COLUMN IF EXISTS mikrotik_object_id;
	`)
	if err != nil {
		return err
	}

	return nil
}
