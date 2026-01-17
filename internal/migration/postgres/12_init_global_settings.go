package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitGlobalSettings, downInitGlobalSettings)
}

func upInitGlobalSettings(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// 1. Global Settings Table
		`CREATE TABLE IF NOT EXISTS global_settings (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			setting_key VARCHAR(100) UNIQUE NOT NULL,
			setting_value TEXT,
			setting_type VARCHAR(50) DEFAULT 'string', -- string, number, boolean, json
			category VARCHAR(50) DEFAULT 'general',
			description TEXT,
			is_public BOOLEAN DEFAULT FALSE,
			created_by UUID REFERENCES users(id),
			updated_by UUID REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		`CREATE INDEX IF NOT EXISTS idx_global_settings_key ON global_settings(setting_key);`,
		`CREATE INDEX IF NOT EXISTS idx_global_settings_category ON global_settings(category);`,

		`DROP TRIGGER IF EXISTS set_updated_at_global_settings ON global_settings;`,
		`CREATE TRIGGER set_updated_at_global_settings
			BEFORE UPDATE ON global_settings
			FOR EACH ROW EXECUTE FUNCTION set_updated_at();`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func downInitGlobalSettings(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP TABLE IF EXISTS global_settings;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
