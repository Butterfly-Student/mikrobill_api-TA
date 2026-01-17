package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"go.uber.org/zap"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
)

type TenantDomain interface {
	CreateTenant(ctx context.Context, input model.CreateTenantRequest, createdBy uuid.UUID) (*model.Tenant, error)
	GetTenant(ctx context.Context, id uuid.UUID) (*model.Tenant, error)
	ListTenants(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, input model.UpdateTenantRequest, updatedBy uuid.UUID) (*model.Tenant, error)
	DeleteTenant(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	GetTenantStats(ctx context.Context, tenantID uuid.UUID) (*model.TenantStatsResponse, error)
	CheckResourceLimit(ctx context.Context, tenantID uuid.UUID, resourceType string) error
}

type tenantDomain struct {
	databasePort outbound_port.DatabasePort
	logger       *zap.Logger
}

// NewTenantDomain creates a new tenant domain
func NewTenantDomain(
	databasePort outbound_port.DatabasePort,
	logger *zap.Logger,
) TenantDomain {
	return &tenantDomain{
		databasePort: databasePort,
		logger:       logger,
	}
}

// CreateTenant creates a new tenant
func (d *tenantDomain) CreateTenant(ctx context.Context, input model.CreateTenantRequest, createdBy uuid.UUID) (*model.Tenant, error) {
	d.logger.Info("Creating new tenant",
		zap.String("name", input.Name),
		zap.String("created_by", createdBy.String()),
	)

	// Create tenant in database
	tenant, err := d.databasePort.Tenant().CreateTenant(ctx, input)
	if err != nil {
		d.logger.Error("Failed to create tenant",
			zap.Error(err),
			zap.String("name", input.Name),
		)
		return nil, stacktrace.Propagate(err, "failed to create tenant")
	}

	d.logger.Info("Tenant created successfully",
		zap.String("tenant_id", tenant.ID),
		zap.String("name", tenant.Name),
	)

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (d *tenantDomain) GetTenant(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	tenant, err := d.databasePort.Tenant().GetByID(ctx, id)
	if err != nil {
		d.logger.Error("Failed to get tenant",
			zap.Error(err),
			zap.String("tenant_id", id.String()),
		)
		return nil, stacktrace.Propagate(err, "tenant not found")
	}

	return tenant, nil
}

// ListTenants lists all tenants (Super Admin only)
func (d *tenantDomain) ListTenants(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error) {
	tenants, err := d.databasePort.Tenant().List(ctx, filter)
	if err != nil {
		d.logger.Error("Failed to list tenants", zap.Error(err))
		return nil, stacktrace.Propagate(err, "failed to list tenants")
	}

	d.logger.Info("Listed tenants", zap.Int("count", len(tenants)))
	return tenants, nil
}

// UpdateTenant updates a tenant
func (d *tenantDomain) UpdateTenant(ctx context.Context, id uuid.UUID, input model.UpdateTenantRequest, updatedBy uuid.UUID) (*model.Tenant, error) {
	// Verify tenant exists
	existing, err := d.GetTenant(ctx, id)
	if err != nil {
		return nil, err
	}

	d.logger.Info("Updating tenant",
		zap.String("tenant_id", id.String()),
		zap.String("updated_by", updatedBy.String()),
	)

	// Update tenant
	// Update tenant
	updated, err := d.databasePort.Tenant().Update(ctx, id, input)
	if err != nil {
		d.logger.Error("Failed to update tenant",
			zap.Error(err),
			zap.String("tenant_id", id.String()),
		)
		return nil, stacktrace.Propagate(err, "failed to update tenant")
	}

	d.logger.Info("Tenant updated successfully",
		zap.String("tenant_id", id.String()),
		zap.String("name", existing.Name),
	)

	return updated, nil
}

// DeleteTenant soft deletes a tenant
func (d *tenantDomain) DeleteTenant(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	// Verify tenant exists
	tenant, err := d.GetTenant(ctx, id)
	if err != nil {
		return err
	}

	d.logger.Warn("Deleting tenant",
		zap.String("tenant_id", id.String()),
		zap.String("name", tenant.Name),
		zap.String("deleted_by", deletedBy.String()),
	)

	// Soft delete tenant
	if err := d.databasePort.Tenant().Delete(ctx, id); err != nil {
		d.logger.Error("Failed to delete tenant",
			zap.Error(err),
			zap.String("tenant_id", id.String()),
		)
		return stacktrace.Propagate(err, "failed to delete tenant")
	}

	d.logger.Info("Tenant deleted successfully",
		zap.String("tenant_id", id.String()),
		zap.String("name", tenant.Name),
	)

	return nil
}

// GetTenantStats retrieves tenant usage statistics
func (d *tenantDomain) GetTenantStats(ctx context.Context, tenantID uuid.UUID) (*model.TenantStatsResponse, error) {
	stats, err := d.databasePort.Tenant().GetStats(ctx, tenantID)
	if err != nil {
		d.logger.Error("Failed to get tenant stats",
			zap.Error(err),
			zap.String("tenant_id", tenantID.String()),
		)
		return nil, stacktrace.Propagate(err, "failed to get tenant stats")
	}

	d.logger.Debug("Retrieved tenant stats",
		zap.String("tenant_id", tenantID.String()),
		zap.Int("mikrotiks", stats.MikrotiksCount),
		zap.Int("users", stats.NetworkUsersCount),
	)

	return stats, nil
}

// CheckResourceLimit checks if tenant is under the limit for a specific resource
func (d *tenantDomain) CheckResourceLimit(ctx context.Context, tenantID uuid.UUID, resourceType string) error {
	// Get tenant
	tenant, err := d.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	// Get current stats
	stats, err := d.GetTenantStats(ctx, tenantID)
	if err != nil {
		return err
	}

	// Check limits
	switch resourceType {
	case "mikrotik":
		if !tenant.IsUnderLimit(resourceType, stats.MikrotiksCount) {
			return fmt.Errorf("mikrotik limit reached (%d/%d)", stats.MikrotiksCount, tenant.MaxMikrotiks)
		}
	case "network_user":
		if !tenant.IsUnderLimit(resourceType, stats.NetworkUsersCount) {
			return fmt.Errorf("network user limit reached (%d/%d)", stats.NetworkUsersCount, tenant.MaxNetworkUsers)
		}
	case "staff_user":
		if !tenant.IsUnderLimit(resourceType, stats.StaffUsersCount) {
			return fmt.Errorf("staff user limit reached (%d/%d)", stats.StaffUsersCount, tenant.MaxStaffUsers)
		}
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	return nil
}

