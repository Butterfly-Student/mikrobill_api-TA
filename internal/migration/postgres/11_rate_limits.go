package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upRateLimits, downRateLimits)
}

func upRateLimits(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// Rate limit rules table
		`CREATE TABLE IF NOT EXISTS rate_limit_rules (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			
			-- Matching criteria
			endpoint_pattern VARCHAR(200),  -- e.g., '/api/customers/*' or '/api/*'
			method VARCHAR(10),              -- GET, POST, etc. NULL = all methods
			user_role user_role,             -- NULL = all roles
			
			-- Limits
			requests_per_minute INTEGER NOT NULL DEFAULT 60,
			requests_per_hour INTEGER NOT NULL DEFAULT 1000,
			requests_per_day INTEGER NOT NULL DEFAULT 10000,
			
			-- Status
			is_active BOOLEAN DEFAULT true,
			priority INTEGER DEFAULT 0,  -- Higher priority rules checked first
			
			-- Audit
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_rate_limit_rules_active ON rate_limit_rules(is_active, priority);`,
		`CREATE INDEX IF NOT EXISTS idx_rate_limit_rules_endpoint ON rate_limit_rules(endpoint_pattern);`,

		`DROP TRIGGER IF EXISTS set_updated_at_rate_limit_rules ON rate_limit_rules;`,
		`CREATE TRIGGER set_updated_at_rate_limit_rules
			BEFORE UPDATE ON rate_limit_rules
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,

		// Rate limit violations log
		`CREATE TABLE IF NOT EXISTS rate_limit_violations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			ip_address INET NOT NULL,
			endpoint VARCHAR(200) NOT NULL,
			method VARCHAR(10),
			rule_name VARCHAR(100),
			exceeded_limit VARCHAR(50),  -- 'per_minute', 'per_hour', 'per_day'
			request_count INTEGER,
			user_agent TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_rate_violations_user ON rate_limit_violations(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_rate_violations_ip ON rate_limit_violations(ip_address);`,
		`CREATE INDEX IF NOT EXISTS idx_rate_violations_created ON rate_limit_violations(created_at);`,

		// Insert default rules
		`INSERT INTO rate_limit_rules (name, description, endpoint_pattern, requests_per_minute, requests_per_hour, requests_per_day, priority) VALUES
		('default_api', 'Default API rate limit', '/api/*', 60, 1000, 10000, 0),
		('auth_endpoints', 'Authentication endpoints', '/api/auth/*', 10, 50, 200, 10),
		('public_endpoints', 'Public endpoints', '/api/public/*', 120, 2000, 20000, 5)
		ON CONFLICT (name) DO NOTHING;`,

		// Role-specific limits
		`INSERT INTO rate_limit_rules (name, description, user_role, requests_per_minute, requests_per_hour, requests_per_day, priority) VALUES
		('super_admin_limit', 'Super admin unlimited', 'SUPER_ADMIN', 10000, 100000, 1000000, 100),
		('tenant_owner_limit', 'Tenant owner high limit', 'TENANT_OWNER', 200, 5000, 50000, 90),
		('tenant_admin_limit', 'Tenant admin limit', 'TENANT_ADMIN', 120, 3000, 30000, 80),
		('tenant_tech_limit', 'Tenant technician limit', 'TENANT_TECHNICIAN', 100, 2000, 20000, 70),
		('tenant_viewer_limit', 'Tenant viewer limit', 'TENANT_VIEWER', 60, 1000, 10000, 60)
		ON CONFLICT (name) DO NOTHING;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downRateLimits(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS rate_limit_violations;`,
		`DROP TABLE IF EXISTS rate_limit_rules;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
