package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upRbac, downRbac)
}

func upRbac(ctx context.Context, tx *sql.Tx) error {
	// Execute the SQL commands for Up migration
	queries := []string{
		`CREATE OR REPLACE FUNCTION set_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		`DO $$ BEGIN
			CREATE TYPE user_role AS ENUM ('superadmin', 'admin', 'technician', 'sales', 'cs', 'finance', 'viewer');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		`DO $$ BEGIN
			CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'banned');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(50) UNIQUE NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			permissions JSONB DEFAULT '[]',
			is_system BOOLEAN DEFAULT false,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,

		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			encrypted_password TEXT NOT NULL,
			fullname VARCHAR(255) NOT NULL,
			phone VARCHAR(20),
			avatar TEXT,
			role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
			user_role user_role NOT NULL DEFAULT 'viewer',
			status user_status NOT NULL DEFAULT 'active',
			last_login TIMESTAMPTZ,
			last_ip INET,
			failed_login_attempts INTEGER DEFAULT 0,
			locked_until TIMESTAMPTZ,
			password_changed_at TIMESTAMPTZ,
			force_password_change BOOLEAN DEFAULT false,
			two_factor_enabled BOOLEAN DEFAULT false,
			two_factor_secret TEXT,
			api_token TEXT UNIQUE,
			api_token_expires_at TIMESTAMPTZ,
			created_by UUID REFERENCES users(id) ON DELETE SET NULL,
			updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users (role_id);`,
		`CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);`,
		`CREATE INDEX IF NOT EXISTS idx_users_api_token ON users (api_token);`,

		`DROP TRIGGER IF EXISTS set_updated_at_users ON users;`,
		`CREATE TRIGGER set_updated_at_users
			BEFORE UPDATE ON users
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,

		`DROP TRIGGER IF EXISTS set_updated_at_roles ON roles;`,
		`CREATE TRIGGER set_updated_at_roles
			BEFORE UPDATE ON roles
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,

		`INSERT INTO roles (name, display_name, description, is_system) VALUES
		('superadmin', 'Super Administrator', 'Full system access', true),
		('admin', 'Administrator', 'Administrative access', true),
		('technician', 'Technician', 'Technical operations', true),
		('sales', 'Sales', 'Sales operations', true),
		('cs', 'Customer Service', 'Customer support', true),
		('finance', 'Finance', 'Financial operations', true),
		('viewer', 'Viewer', 'Read-only access', true)
		ON CONFLICT (name) DO NOTHING;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downRbac(ctx context.Context, tx *sql.Tx) error {
	// Execute the SQL commands for Down migration
	queries := []string{
		`DROP TABLE IF EXISTS users;`,
		`DROP TABLE IF EXISTS roles;`,
		`DROP TYPE IF EXISTS user_status;`,
		`DROP TYPE IF EXISTS user_role;`,
		`DROP FUNCTION IF EXISTS set_updated_at;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
