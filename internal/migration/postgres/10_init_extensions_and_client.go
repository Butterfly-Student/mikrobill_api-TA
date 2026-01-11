package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitExtensionsAndClient, downInitExtensionsAndClient)
}

func upInitExtensionsAndClient(ctx context.Context, tx *sql.Tx) error {
	// Create uuid-ossp extension
	_, err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		return err
	}

	// Create ENUM types if they don't exist
	queries := []string{
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'mikrotik_status') THEN CREATE TYPE mikrotik_status AS ENUM ('offline', 'online', 'error'); END IF; END $$;`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'profile_type') THEN CREATE TYPE profile_type AS ENUM ('pppoe', 'hotspot', 'static_ip'); END IF; END $$;`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'customer_status') THEN CREATE TYPE customer_status AS ENUM ('active', 'suspended', 'inactive', 'pending'); END IF; END $$;`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'service_status') THEN CREATE TYPE service_status AS ENUM ('active', 'suspended', 'terminated'); END IF; END $$;`,
	}

	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}

	// Create clients table
	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS clients (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100),
		bearer_key VARCHAR(255) UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);`)
	if err != nil {
		return err
	}

	return nil
}

func downInitExtensionsAndClient(ctx context.Context, tx *sql.Tx) error {
	// Drop clients table
	_, err := tx.Exec(`DROP TABLE IF EXISTS clients;`)
	if err != nil {
		return err
	}

	// Drop ENUM types
	queries := []string{
		`DROP TYPE IF EXISTS service_status;`,
		`DROP TYPE IF EXISTS customer_status;`,
		`DROP TYPE IF EXISTS profile_type;`,
		`DROP TYPE IF EXISTS mikrotik_status;`,
	}

	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}

	return nil
}
