package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitRbac, downInitRbac)
}

func upInitRbac(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// 1. Roles Table
		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
			name VARCHAR(50) NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			permissions JSONB DEFAULT '[]',
			is_system BOOLEAN DEFAULT false,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (tenant_id, name)
		);`,

		// Handle NULL tenant_id for global roles (system roles)
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_global_name ON roles (name) WHERE tenant_id IS NULL;`,

		`DROP TRIGGER IF EXISTS set_updated_at_roles ON roles;`,
		`CREATE TRIGGER set_updated_at_roles
			BEFORE UPDATE ON roles
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,

		// 2. Users Table
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE, -- Primary/Default tenant
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			encrypted_password TEXT NOT NULL,
			fullname VARCHAR(255) NOT NULL,
			phone VARCHAR(20),
			avatar TEXT,
			
			role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
			user_role user_role NOT NULL DEFAULT 'viewer',
			status user_status NOT NULL DEFAULT 'active',
			
			-- Security
			is_superadmin BOOLEAN DEFAULT FALSE,
			last_login_at TIMESTAMPTZ,
			last_ip INET,
			failed_login_attempts INTEGER DEFAULT 0,
			locked_until TIMESTAMPTZ,
			password_changed_at TIMESTAMPTZ,
			force_password_change BOOLEAN DEFAULT false,
			two_factor_enabled BOOLEAN DEFAULT false,
			two_factor_secret TEXT,
			
			-- Audit
			created_by UUID, -- References users(id), done later via ALTER
			updated_by UUID, -- References users(id), done later via ALTER
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ
		);`,

		`CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
		`CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);`,
		`CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);`,

		`DROP TRIGGER IF EXISTS set_updated_at_users ON users;`,
		`CREATE TRIGGER set_updated_at_users
			BEFORE UPDATE ON users
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,

		// 3. Add FK constraints for users (created_by, updated_by)
		`ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_created_by;`,
		`ALTER TABLE users ADD CONSTRAINT fk_users_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;`,
		`ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_updated_by;`,
		`ALTER TABLE users ADD CONSTRAINT fk_users_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL;`,

		// 4. User Sessions
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL UNIQUE,
			ip_address INET,
			user_agent TEXT,
			is_active BOOLEAN DEFAULT true,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_sessions_user ON user_sessions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON user_sessions(token_hash);`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downInitRbac(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS user_sessions;`,
		`DROP TABLE IF EXISTS users;`,
		`DROP TABLE IF EXISTS roles;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
