package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upEnhanceSessions, downEnhanceSessions)
}

func upEnhanceSessions(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// Drop existing sessions table to rebuild with enhancements
		`DROP TABLE IF EXISTS user_sessions CASCADE;`,

		// Recreate with Redis-compatible structure
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
			
			-- Token management
			token_hash TEXT NOT NULL UNIQUE,
			refresh_token_hash TEXT UNIQUE,
			
			-- Session metadata
			ip_address INET,
			user_agent TEXT,
			device_info JSONB DEFAULT '{}',
			
			-- Status
			is_active BOOLEAN DEFAULT true,
			expires_at TIMESTAMPTZ NOT NULL,
			refreshed_at TIMESTAMPTZ,
			last_activity_at TIMESTAMPTZ DEFAULT now(),
			
			-- Audit
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			revoked_at TIMESTAMPTZ,
			revoked_by UUID REFERENCES users(id) ON DELETE SET NULL,
			revoke_reason TEXT
		);`,

		// Indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_sessions_user ON user_sessions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_tenant ON user_sessions(user_id, tenant_id);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON user_sessions(token_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON user_sessions(refresh_token_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_active ON user_sessions(is_active, expires_at);`,

		// Session activity tracking
		`CREATE TABLE IF NOT EXISTS session_activities (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			session_id UUID NOT NULL REFERENCES user_sessions(id) ON DELETE CASCADE,
			action VARCHAR(100) NOT NULL,  -- login, refresh, logout, api_call, etc.
			ip_address INET,
			user_agent TEXT,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_session_activities_session ON session_activities(session_id);`,
		`CREATE INDEX IF NOT EXISTS idx_session_activities_created ON session_activities(created_at);`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downEnhanceSessions(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS session_activities;`,
		`DROP TABLE IF EXISTS user_sessions;`,

		// Recreate simple version
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
