package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateProfileSpecificTables, downCreateProfileSpecificTables)
}

func upCreateProfileSpecificTables(ctx context.Context, tx *sql.Tx) error {
	// Create profile-specific tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS mikrotik_profile_pppoe (
			profile_id UUID PRIMARY KEY REFERENCES mikrotik_profiles(id) ON DELETE CASCADE,
			local_address INET,
			remote_address INET,
			address_pool VARCHAR(50),
			mtu INTEGER DEFAULT 1480,
			mru INTEGER DEFAULT 1480,
			use_mpls BOOLEAN DEFAULT false,
			use_compression BOOLEAN DEFAULT false,
			use_encryption BOOLEAN DEFAULT false
		);`,

		`CREATE TABLE IF NOT EXISTS mikrotik_profile_hotspot (
			profile_id UUID PRIMARY KEY REFERENCES mikrotik_profiles(id) ON DELETE CASCADE,
			shared_users INTEGER DEFAULT 1,
			address_pool VARCHAR(50),
			mac_auth BOOLEAN DEFAULT false,
			mac_auth_mode VARCHAR(20) DEFAULT 'none',
			login_timeout_seconds INTEGER,
			cookie_timeout_seconds INTEGER
		);`,

		`CREATE TABLE IF NOT EXISTS mikrotik_profile_static_ip (
			profile_id UUID PRIMARY KEY REFERENCES mikrotik_profiles(id) ON DELETE CASCADE,
			ip_address INET NOT NULL,
			gateway INET NOT NULL,
			vlan_id INTEGER,
			routing_mark VARCHAR(50)
		);`,
	}

	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func downCreateProfileSpecificTables(ctx context.Context, tx *sql.Tx) error {
	// Drop tables in reverse order
	tables := []string{
		"mikrotik_profile_static_ip",
		"mikrotik_profile_hotspot",
		"mikrotik_profile_pppoe",
	}

	for _, table := range tables {
		if _, err := tx.Exec(`DROP TABLE IF EXISTS ` + table + `;`); err != nil {
			return err
		}
	}
	return nil
}
