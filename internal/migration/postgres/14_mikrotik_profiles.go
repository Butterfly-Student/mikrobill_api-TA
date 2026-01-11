package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateMikrotikProfilesTable, downCreateMikrotikProfilesTable)
}

func upCreateMikrotikProfilesTable(ctx context.Context, tx *sql.Tx) error {
	// Create mikrotik_profiles table
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS mikrotik_profiles (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			profile_type profile_type NOT NULL,
			mikrotik_object_id VARCHAR(50) NOT NULL,
			rate_limit_up_kbps INTEGER,
			rate_limit_down_kbps INTEGER,
			idle_timeout_seconds INTEGER,
			session_timeout_seconds INTEGER,
			keepalive_timeout_seconds INTEGER,
			only_one BOOLEAN NOT NULL DEFAULT false,
			status_authentication BOOLEAN NOT NULL DEFAULT true,
			dns_server INET,
			is_active BOOLEAN NOT NULL DEFAULT true,
			sync_with_mikrotik BOOLEAN NOT NULL DEFAULT true,
			last_sync TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (mikrotik_id, mikrotik_object_id),
			UNIQUE (mikrotik_id, name, profile_type)
		);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downCreateMikrotikProfilesTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS mikrotik_profiles;`)
	if err != nil {
		return err
	}
	return nil
}
