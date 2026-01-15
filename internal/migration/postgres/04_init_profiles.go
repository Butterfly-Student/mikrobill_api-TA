package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitProfiles, downInitProfiles)
}

func upInitProfiles(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// 1. Main Profiles Table
		`CREATE TABLE IF NOT EXISTS mikrotik_profiles (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			mikrotik_id UUID NOT NULL REFERENCES mikrotik(id) ON DELETE CASCADE,
			
			name VARCHAR(100) NOT NULL,
			type profile_type NOT NULL,
			is_default BOOLEAN DEFAULT false,
			
			-- Common parameters
			rate_limit VARCHAR(100), -- [rx-bits]/[tx-bits] [rx-burst-bits]/[tx-burst-bits] ...
			session_timeout INTERVAL,
			idle_timeout INTERVAL,
			
			-- Billing parameters
			price NUMERIC(15,2) DEFAULT 0.00,
			tax_rate NUMERIC(5,2) DEFAULT 0.00,
			
			-- Metadata & Audit
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ,
			
			UNIQUE (mikrotik_id, name, deleted_at)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_profiles_tenant ON mikrotik_profiles(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_profiles_mikrotik ON mikrotik_profiles(mikrotik_id);`,
		`CREATE INDEX IF NOT EXISTS idx_profiles_deleted_at ON mikrotik_profiles(deleted_at);`,

		`DROP TRIGGER IF EXISTS set_updated_at_profiles ON mikrotik_profiles;`,
		`CREATE TRIGGER set_updated_at_profiles
			BEFORE UPDATE ON mikrotik_profiles
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,

		// 2. PPPoE Specific
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

		// 3. Hotspot Specific
		`CREATE TABLE IF NOT EXISTS mikrotik_profile_hotspot (
			profile_id UUID PRIMARY KEY REFERENCES mikrotik_profiles(id) ON DELETE CASCADE,
			shared_users INTEGER DEFAULT 1,
			address_pool VARCHAR(50),
			mac_auth BOOLEAN DEFAULT false,
			mac_auth_mode VARCHAR(20) DEFAULT 'none',
			login_timeout_seconds INTEGER,
			cookie_timeout_seconds INTEGER
		);`,

		// 4. Static IP Specific
		`CREATE TABLE IF NOT EXISTS mikrotik_profile_static_ip (
			profile_id UUID PRIMARY KEY REFERENCES mikrotik_profiles(id) ON DELETE CASCADE,
			ip_address INET,
			gateway INET,
			vlan_id INTEGER,
			routing_mark VARCHAR(50)
		);`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downInitProfiles(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS mikrotik_profile_static_ip;`,
		`DROP TABLE IF EXISTS mikrotik_profile_hotspot;`,
		`DROP TABLE IF EXISTS mikrotik_profile_pppoe;`,
		`DROP TABLE IF EXISTS mikrotik_profiles;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
