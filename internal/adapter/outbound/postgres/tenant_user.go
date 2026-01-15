package postgres_outbound_adapter

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	outbound_port "prabogo/internal/port/outbound"
)

const tableTenantUsers = "tenant_users"

type tenantUserAdapter struct {
	db outbound_port.DatabaseExecutor
}

func NewTenantUserAdapter(db outbound_port.DatabaseExecutor) outbound_port.TenantUserDatabasePort {
	return &tenantUserAdapter{db: db}
}

// HasAccess checks if user has access to the specified tenant
func (a *tenantUserAdapter) HasAccess(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	query, _, err := goqu.Dialect("postgres").
		From(tableTenantUsers).
		Select(goqu.COUNT("*")).
		Where(goqu.Ex{
			"user_id":   userID,
			"tenant_id": tenantID,
			"is_active": true,
		}).
		ToSQL()

	if err != nil {
		return false, stacktrace.Propagate(err, "failed to build has access query")
	}

	var count int
	err = a.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return false, stacktrace.Propagate(err, "failed to check tenant access")
	}

	return count > 0, nil
}

// GetPrimaryTenant retrieves the primary tenant for a user
func (a *tenantUserAdapter) GetPrimaryTenant(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	query, _, err := goqu.Dialect("postgres").
		From(tableTenantUsers).
		Select("tenant_id").
		Where(goqu.Ex{
			"user_id":    userID,
			"is_primary": true,
			"is_active":  true,
		}).
		Limit(1).
		ToSQL()

	if err != nil {
		return uuid.Nil, stacktrace.Propagate(err, "failed to build get primary tenant query")
	}

	var tenantID uuid.UUID
	err = a.db.QueryRowContext(ctx, query).Scan(&tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No primary tenant found, try to get any active tenant
			return a.getAnyActiveTenant(ctx, userID)
		}
		return uuid.Nil, stacktrace.Propagate(err, "failed to get primary tenant")
	}

	return tenantID, nil
}

// getAnyActiveTenant retrieves any active tenant for a user
func (a *tenantUserAdapter) getAnyActiveTenant(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	query, _, err := goqu.Dialect("postgres").
		From(tableTenantUsers).
		Select("tenant_id").
		Where(goqu.Ex{
			"user_id":   userID,
			"is_active": true,
		}).
		Order(goqu.C("created_at").Asc()).
		Limit(1).
		ToSQL()

	if err != nil {
		return uuid.Nil, stacktrace.Propagate(err, "failed to build get any tenant query")
	}

	var tenantID uuid.UUID
	err = a.db.QueryRowContext(ctx, query).Scan(&tenantID)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, stacktrace.NewError("user has no associated tenants")
		}
		return uuid.Nil, stacktrace.Propagate(err, "failed to get any active tenant")
	}

	return tenantID, nil
}

// GetTenantsForUser retrieves all tenants for a user
func (a *tenantUserAdapter) GetTenantsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query, _, err := goqu.Dialect("postgres").
		From(tableTenantUsers).
		Select("tenant_id").
		Where(goqu.Ex{
			"user_id":   userID,
			"is_active": true,
		}).
		ToSQL()

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build get tenants query")
	}

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenants for user")
	}
	defer rows.Close()

	var tenants []uuid.UUID
	for rows.Next() {
		var tenantID uuid.UUID
		if err := rows.Scan(&tenantID); err != nil {
			return nil, stacktrace.Propagate(err, "failed to scan tenant ID")
		}
		tenants = append(tenants, tenantID)
	}

	return tenants, nil
}
