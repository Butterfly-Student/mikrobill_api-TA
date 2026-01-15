package postgres_outbound_adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	outbound_port "prabogo/internal/port/outbound"
)

const tableTenants = "tenants"

type tenantAdapter struct {
	db outbound_port.DatabaseExecutor
}

func NewTenantAdapter(db outbound_port.DatabaseExecutor) outbound_port.TenantDatabasePort {
	return &tenantAdapter{db: db}
}

// CreateTenant creates a new tenant
func (a *tenantAdapter) CreateTenant(ctx context.Context, input model.TenantInput) (*model.Tenant, error) {
	tenant := &model.Tenant{
		ID:              uuid.New(),
		Name:            input.Name,
		Timezone:        "Asia/Jakarta",
		IsActive:        true,
		Status:          "active",
		MaxMikrotiks:    3,
		MaxNetworkUsers: 50,
		MaxStaffUsers:   5,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Apply optional fields
	if input.Subdomain != nil {
		tenant.Subdomain = input.Subdomain
	}
	if input.CompanyName != nil {
		tenant.CompanyName = input.CompanyName
	}
	if input.Phone != nil {
		tenant.Phone = input.Phone
	}
	if input.Address != nil {
		tenant.Address = input.Address
	}
	if input.Timezone != nil {
		tenant.Timezone = *input.Timezone
	}
	if input.MaxMikrotiks != nil {
		tenant.MaxMikrotiks = *input.MaxMikrotiks
	}
	if input.MaxNetworkUsers != nil {
		tenant.MaxNetworkUsers = *input.MaxNetworkUsers
	}
	if input.MaxStaffUsers != nil {
		tenant.MaxStaffUsers = *input.MaxStaffUsers
	}

	record := goqu.Record{
		"id":                tenant.ID,
		"name":              tenant.Name,
		"subdomain":         tenant.Subdomain,
		"company_name":      tenant.CompanyName,
		"phone":             tenant.Phone,
		"address":           tenant.Address,
		"timezone":          tenant.Timezone,
		"is_active":         tenant.IsActive,
		"status":            tenant.Status,
		"max_mikrotiks":     tenant.MaxMikrotiks,
		"max_network_users": tenant.MaxNetworkUsers,
		"max_staff_users":   tenant.MaxStaffUsers,
		"features":          `{"api_access": true, "reports": true, "backup": true}`,
		"metadata":          `{}`,
		"created_at":        tenant.CreatedAt,
		"updated_at":        tenant.UpdatedAt,
	}

	query, _, err := goqu.Dialect("postgres").
		Insert(tableTenants).
		Rows(record).
		Returning("*").
		ToSQL()

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build insert tenant query")
	}

	var result model.Tenant
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.Name,
		&result.Subdomain,
		&result.CompanyName,
		&result.Phone,
		&result.Address,
		&result.Timezone,
		&result.IsActive,
		&result.Status,
		&result.MaxMikrotiks,
		&result.MaxNetworkUsers,
		&result.MaxStaffUsers,
		&result.Features,
		&result.Metadata,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.SuspendedAt,
		&result.DeletedAt,
	)

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create tenant")
	}

	return &result, nil
}

// GetByID retrieves a tenant by ID
func (a *tenantAdapter) GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	query, _, err := goqu.Dialect("postgres").
		From(tableTenants).
		Where(goqu.Ex{"id": id, "deleted_at": nil}).
		ToSQL()

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build get tenant query")
	}

	var result model.Tenant
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.Name,
		&result.Subdomain,
		&result.CompanyName,
		&result.Phone,
		&result.Address,
		&result.Timezone,
		&result.IsActive,
		&result.Status,
		&result.MaxMikrotiks,
		&result.MaxNetworkUsers,
		&result.MaxStaffUsers,
		&result.Features,
		&result.Metadata,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.SuspendedAt,
		&result.DeletedAt,
	)

	if err != nil {
		return nil, stacktrace.Propagate(err, "tenant not found")
	}

	return &result, nil
}

// List retrieves all tenants (Super Admin only)
func (a *tenantAdapter) List(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error) {
	ds := goqu.Dialect("postgres").
		From(tableTenants).
		Where(goqu.Ex{"deleted_at": nil})

	// Apply filters
	if filter.Status != nil {
		ds = ds.Where(goqu.Ex{"status": *filter.Status})
	}
	if filter.IsActive != nil {
		ds = ds.Where(goqu.Ex{"is_active": *filter.IsActive})
	}
	if filter.Search != nil && *filter.Search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", *filter.Search)
		ds = ds.Where(goqu.Or(
			goqu.C("name").ILike(searchPattern),
			goqu.C("company_name").ILike(searchPattern),
		))
	}

	ds = ds.Order(goqu.C("created_at").Desc())

	if filter.Limit > 0 {
		ds = ds.Limit(uint(filter.Limit))
	}
	if filter.Offset > 0 {
		ds = ds.Offset(uint(filter.Offset))
	}

	query, _, err := ds.ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build list tenants query")
	}

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list tenants")
	}
	defer rows.Close()

	var tenants []model.Tenant
	for rows.Next() {
		var t model.Tenant
		err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.Subdomain,
			&t.CompanyName,
			&t.Phone,
			&t.Address,
			&t.Timezone,
			&t.IsActive,
			&t.Status,
			&t.MaxMikrotiks,
			&t.MaxNetworkUsers,
			&t.MaxStaffUsers,
			&t.Features,
			&t.Metadata,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.SuspendedAt,
			&t.DeletedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to scan tenant")
		}
		tenants = append(tenants, t)
	}

	return tenants, nil
}

// Update updates a tenant
func (a *tenantAdapter) Update(ctx context.Context, id uuid.UUID, input model.TenantInput) error {
	record := goqu.Record{
		"updated_at": time.Now(),
	}

	if input.Name != "" {
		record["name"] = input.Name
	}
	if input.Subdomain != nil {
		record["subdomain"] = input.Subdomain
	}
	if input.CompanyName != nil {
		record["company_name"] = input.CompanyName
	}
	if input.Phone != nil {
		record["phone"] = input.Phone
	}
	if input.Address != nil {
		record["address"] = input.Address
	}
	if input.Timezone != nil {
		record["timezone"] = input.Timezone
	}
	if input.IsActive != nil {
		record["is_active"] = input.IsActive
	}
	if input.Status != nil {
		record["status"] = input.Status
	}
	if input.MaxMikrotiks != nil {
		record["max_mikrotiks"] = input.MaxMikrotiks
	}
	if input.MaxNetworkUsers != nil {
		record["max_network_users"] = input.MaxNetworkUsers
	}
	if input.MaxStaffUsers != nil {
		record["max_staff_users"] = input.MaxStaffUsers
	}

	query, _, err := goqu.Dialect("postgres").
		Update(tableTenants).
		Set(record).
		Where(goqu.Ex{"id": id, "deleted_at": nil}).
		ToSQL()

	if err != nil {
		return stacktrace.Propagate(err, "failed to build update tenant query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update tenant")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return stacktrace.NewError("tenant not found")
	}

	return nil
}

// Delete soft deletes a tenant
func (a *tenantAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	query, _, err := goqu.Dialect("postgres").
		Update(tableTenants).
		Set(goqu.Record{"deleted_at": time.Now()}).
		Where(goqu.Ex{"id": id, "deleted_at": nil}).
		ToSQL()

	if err != nil {
		return stacktrace.Propagate(err, "failed to build delete tenant query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete tenant")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return stacktrace.NewError("tenant not found")
	}

	return nil
}

// GetStats retrieves tenant stats
func (a *tenantAdapter) GetStats(ctx context.Context, tenantID uuid.UUID) (*model.TenantStats, error) {
	stats := &model.TenantStats{
		TenantID: tenantID,
	}

	// Get tenant limits first
	tenant, err := a.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	stats.MaxMikrotiks = tenant.MaxMikrotiks
	stats.MaxNetworkUsers = tenant.MaxNetworkUsers
	stats.MaxStaffUsers = tenant.MaxStaffUsers

	// Note: Count queries would need to be implemented separately
	// For now, return empty stats
	stats.MikrotiksCount = 0
	stats.NetworkUsersCount = 0
	stats.StaffUsersCount = 0

	return stats, nil
}
