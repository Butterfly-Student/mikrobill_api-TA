package postgres_outbound_adapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"gorm.io/gorm"

	outbound_port "MikrOps/internal/port/outbound"
)

const tableTenantUsers = "tenant_users"

type tenantUserAdapter struct {
	db *gorm.DB
}

func NewTenantUserAdapter(db *gorm.DB) outbound_port.TenantUserDatabasePort {
	return &tenantUserAdapter{db: db}
}

// HasAccess checks if user has access to the specified tenant
func (a *tenantUserAdapter) HasAccess(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	var count int64
	err := a.db.WithContext(ctx).
		Table(tableTenantUsers).
		Where("user_id = ? AND tenant_id = ? AND is_active = ?", userID, tenantID, true).
		Count(&count).Error

	if err != nil {
		return false, stacktrace.Propagate(err, "failed to check tenant access")
	}

	return count > 0, nil
}

// GetPrimaryTenant retrieves the primary tenant for a user
func (a *tenantUserAdapter) GetPrimaryTenant(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var tenantID uuid.UUID
	err := a.db.WithContext(ctx).
		Table(tableTenantUsers).
		Select("tenant_id").
		Where("user_id = ? AND is_primary = ? AND is_active = ?", userID, true, true).
		Limit(1).
		Row().
		Scan(&tenantID)

	if err == gorm.ErrRecordNotFound {
		// No primary tenant found, try to get any active tenant
		return a.getAnyActiveTenant(ctx, userID)
	}

	if err != nil {
		return uuid.Nil, stacktrace.Propagate(err, "failed to get primary tenant")
	}

	return tenantID, nil
}

// getAnyActiveTenant retrieves any active tenant for a user
func (a *tenantUserAdapter) getAnyActiveTenant(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var tenantID uuid.UUID
	err := a.db.WithContext(ctx).
		Table(tableTenantUsers).
		Select("tenant_id").
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("created_at ASC").
		Limit(1).
		Row().
		Scan(&tenantID)

	if err == gorm.ErrRecordNotFound {
		return uuid.Nil, stacktrace.NewError("user has no associated tenants")
	}

	if err != nil {
		return uuid.Nil, stacktrace.Propagate(err, "failed to get any active tenant")
	}

	return tenantID, nil
}

// GetTenantsForUser retrieves all tenants for a user
func (a *tenantUserAdapter) GetTenantsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var tenants []uuid.UUID
	err := a.db.WithContext(ctx).
		Table(tableTenantUsers).
		Select("tenant_id").
		Where("user_id = ? AND is_active = ?", userID, true).
		Scan(&tenants).Error

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenants for user")
	}

	return tenants, nil
}

// AssignUserToTenant assigns a user to a tenant with optional role
func (a *tenantUserAdapter) AssignUserToTenant(ctx context.Context, userID, tenantID uuid.UUID, roleID *uuid.UUID, isPrimary bool) error {
	// Check if assignment already exists
	var count int64
	err := a.db.WithContext(ctx).
		Table(tableTenantUsers).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Count(&count).Error

	if err != nil {
		return stacktrace.Propagate(err, "failed to check existing tenant assignment")
	}

	// If exists, update it
	if count > 0 {
		updates := map[string]interface{}{
			"is_active":  true,
			"is_primary": isPrimary,
		}
		if roleID != nil {
			updates["role_id"] = roleID
		}

		if err := a.db.WithContext(ctx).
			Table(tableTenantUsers).
			Where("user_id = ? AND tenant_id = ?", userID, tenantID).
			Updates(updates).Error; err != nil {
			return stacktrace.Propagate(err, "failed to update tenant assignment")
		}
		return nil
	}

	// Otherwise create new assignment
	insertData := map[string]interface{}{
		"user_id":    userID,
		"tenant_id":  tenantID,
		"is_active":  true,
		"is_primary": isPrimary,
	}
	if roleID != nil {
		insertData["role_id"] = roleID
	}

	if err := a.db.WithContext(ctx).
		Table(tableTenantUsers).
		Create(insertData).Error; err != nil {
		return stacktrace.Propagate(err, "failed to create tenant assignment")
	}

	return nil
}
