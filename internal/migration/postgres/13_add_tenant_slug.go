package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddTenantSlug, downAddTenantSlug)
}

func upAddTenantSlug(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// Add slug column to tenants table
		`ALTER TABLE tenants ADD COLUMN IF NOT EXISTS slug VARCHAR(100);`,

		// Add unique constraint
		`ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_slug_key;`,
		`ALTER TABLE tenants ADD CONSTRAINT tenants_slug_key UNIQUE (slug);`,

		// Add index for performance
		`CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);`,

		// Optional: Generate slug from subdomain or name for existing tenants
		// You can uncomment this if you want auto-generated slugs
		/*
			`UPDATE tenants
			 SET slug = LOWER(REGEXP_REPLACE(COALESCE(subdomain, name), '[^a-zA-Z0-9]', '', 'g'))
			 WHERE slug IS NULL;`,
		*/
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downAddTenantSlug(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		`DROP INDEX IF EXISTS idx_tenants_slug;`,
		`ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_slug_key;`,
		`ALTER TABLE tenants DROP COLUMN IF EXISTS slug;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
