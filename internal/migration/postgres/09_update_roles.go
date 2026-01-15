package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upUpdateRoles, downUpdateRoles)
}

func upUpdateRoles(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// Rename old enum
		`ALTER TYPE user_role RENAME TO user_role_old;`,

		// Create new enum with updated values
		`CREATE TYPE user_role AS ENUM (
			'SUPER_ADMIN',
			'TENANT_OWNER',
			'TENANT_ADMIN',
			'TENANT_TECHNICIAN',
			'TENANT_VIEWER'
		);`,

		// Drop default before altering column type
		`ALTER TABLE users ALTER COLUMN user_role DROP DEFAULT;`,

		// Update users table - migrate old values to new
		`ALTER TABLE users 
		ALTER COLUMN user_role TYPE user_role 
		USING (
			CASE user_role::text
				WHEN 'superadmin' THEN 'SUPER_ADMIN'::user_role
				WHEN 'admin' THEN 'TENANT_ADMIN'::user_role
				WHEN 'technician' THEN 'TENANT_TECHNICIAN'::user_role
				WHEN 'viewer' THEN 'TENANT_VIEWER'::user_role
				ELSE 'TENANT_VIEWER'::user_role
			END
		);`,

		// Drop old enum
		`DROP TYPE user_role_old;`,

		// Update default value
		`ALTER TABLE users ALTER COLUMN user_role SET DEFAULT 'TENANT_VIEWER';`,

		// Update existing superadmin user
		`UPDATE users SET user_role = 'SUPER_ADMIN' WHERE is_superadmin = true;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func downUpdateRoles(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// Rename current enum
		`ALTER TYPE user_role RENAME TO user_role_new;`,

		// Recreate old enum
		`CREATE TYPE user_role AS ENUM ('superadmin', 'admin', 'technician', 'sales', 'cs', 'finance', 'viewer');`,

		// Migrate back to old values
		`ALTER TABLE users 
		ALTER COLUMN user_role TYPE user_role 
		USING (
			CASE user_role::text
				WHEN 'SUPER_ADMIN' THEN 'superadmin'::user_role
				WHEN 'TENANT_OWNER' THEN 'admin'::user_role
				WHEN 'TENANT_ADMIN' THEN 'admin'::user_role
				WHEN 'TENANT_TECHNICIAN' THEN 'technician'::user_role
				WHEN 'TENANT_VIEWER' THEN 'viewer'::user_role
				ELSE 'viewer'::user_role
			END
		);`,

		// Drop new enum
		`DROP TYPE user_role_new;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}
