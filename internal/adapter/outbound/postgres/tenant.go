package postgres_outbound_adapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
)

type tenantAdapter struct {
	db *gorm.DB
}

func NewTenantAdapter(db *gorm.DB) outbound_port.TenantDatabasePort {
	return &tenantAdapter{db: db}
}

// CreateTenant creates a new tenant
func (a *tenantAdapter) CreateTenant(ctx context.Context, req model.CreateTenantRequest) (*model.Tenant, error) {
	tenant := &model.Tenant{
		Name:            req.Name,
		Slug:            req.Slug,
		Subdomain:       req.Subdomain,
		CompanyName:     req.CompanyName,
		Phone:           req.Phone,
		Address:         req.Address,
		Timezone:        "Asia/Jakarta",
		IsActive:        true,
		Status:          "active",
		MaxMikrotiks:    3,
		MaxNetworkUsers: 50,
		MaxStaffUsers:   5,
	}

	// Apply optional fields
	if req.Timezone != nil {
		tenant.Timezone = *req.Timezone
	}
	if req.MaxMikrotiks != nil {
		tenant.MaxMikrotiks = *req.MaxMikrotiks
	}
	if req.MaxNetworkUsers != nil {
		tenant.MaxNetworkUsers = *req.MaxNetworkUsers
	}
	if req.MaxStaffUsers != nil {
		tenant.MaxStaffUsers = *req.MaxStaffUsers
	}
	if req.Features != nil {
		tenant.Features = req.Features
	}
	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}

	if err := a.db.WithContext(ctx).Create(tenant).Error; err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

// GetByID retrieves a tenant by ID
func (a *tenantAdapter) GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	var tenant model.Tenant

	err := a.db.WithContext(ctx).
		Where("id = ?", id).
		First(&tenant).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

// GetBySubdomain retrieves a tenant by subdomain
func (a *tenantAdapter) GetBySubdomain(ctx context.Context, subdomain string) (*model.Tenant, error) {
	var tenant model.Tenant

	err := a.db.WithContext(ctx).
		Where("subdomain = ?", subdomain).
		First(&tenant).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

// GetBySlug retrieves a tenant by slug
func (a *tenantAdapter) GetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	var tenant model.Tenant

	err := a.db.WithContext(ctx).
		Where("slug = ? AND deleted_at IS NULL", slug).
		First(&tenant).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Return nil instead of error for not found case
		}
		return nil, fmt.Errorf("failed to get tenant by slug: %w", err)
	}

	return &tenant, nil
}

// List retrieves all tenants with filtering and pagination
func (a *tenantAdapter) List(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error) {
	var tenants []model.Tenant

	query := a.db.WithContext(ctx).Model(&model.Tenant{})

	// Apply filters
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != nil && *filter.Search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", *filter.Search)
		query = query.Where(
			"name ILIKE ? OR company_name ILIKE ?",
			searchPattern, searchPattern,
		)
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC")

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Execute query
	if err := query.Find(&tenants).Error; err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, nil
}

// Update updates a tenant
func (a *tenantAdapter) Update(ctx context.Context, id uuid.UUID, req model.UpdateTenantRequest) (*model.Tenant, error) {
	updates := make(map[string]interface{})

	// Build update map
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.Subdomain != nil {
		updates["subdomain"] = *req.Subdomain
	}
	if req.CompanyName != nil {
		updates["company_name"] = *req.CompanyName
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.MaxMikrotiks != nil {
		updates["max_mikrotiks"] = *req.MaxMikrotiks
	}
	if req.MaxNetworkUsers != nil {
		updates["max_network_users"] = *req.MaxNetworkUsers
	}
	if req.MaxStaffUsers != nil {
		updates["max_staff_users"] = *req.MaxStaffUsers
	}
	if req.Features != nil {
		updates["features"] = req.Features
	}
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}

	// Execute update
	result := a.db.WithContext(ctx).
		Model(&model.Tenant{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("tenant not found")
	}

	// Fetch and return updated tenant
	return a.GetByID(ctx, id)
}

// Delete soft deletes a tenant
func (a *tenantAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	result := a.db.WithContext(ctx).
		Delete(&model.Tenant{}, id)

	if result.Error != nil {
		return fmt.Errorf("failed to delete tenant: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// Suspend suspends a tenant
func (a *tenantAdapter) Suspend(ctx context.Context, id uuid.UUID) error {
	result := a.db.WithContext(ctx).
		Model(&model.Tenant{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "suspended",
			"is_active":    false,
			"suspended_at": gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to suspend tenant: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// Activate activates a suspended tenant
func (a *tenantAdapter) Activate(ctx context.Context, id uuid.UUID) error {
	result := a.db.WithContext(ctx).
		Model(&model.Tenant{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "active",
			"is_active":    true,
			"suspended_at": nil,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to activate tenant: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// GetStats retrieves tenant usage statistics
func (a *tenantAdapter) GetStats(ctx context.Context, tenantID uuid.UUID) (*model.TenantStatsResponse, error) {
	// Get tenant to retrieve limits
	tenant, err := a.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	stats := &model.TenantStatsResponse{
		TenantID:        tenantID,
		MaxMikrotiks:    tenant.MaxMikrotiks,
		MaxNetworkUsers: tenant.MaxNetworkUsers,
		MaxStaffUsers:   tenant.MaxStaffUsers,
	}

	// Count mikrotiks
	var mikrotiksCount int64
	if err := a.db.WithContext(ctx).
		Model(&model.Mikrotik{}).
		Where("tenant_id = ?", tenantID).
		Count(&mikrotiksCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count mikrotiks: %w", err)
	}
	stats.MikrotiksCount = int(mikrotiksCount)

	// Count network users (customers)
	var networkUsersCount int64
	if err := a.db.WithContext(ctx).
		Model(&model.Customer{}).
		Where("tenant_id = ?", tenantID).
		Count(&networkUsersCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count network users: %w", err)
	}
	stats.NetworkUsersCount = int(networkUsersCount)

	// Count staff users
	var staffUsersCount int64
	if err := a.db.WithContext(ctx).
		Model(&model.TenantUser{}).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Count(&staffUsersCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count staff users: %w", err)
	}
	stats.StaffUsersCount = int(staffUsersCount)

	// Calculate usage percentages
	if stats.MaxMikrotiks > 0 {
		stats.MikrotiksUsagePercent = float64(stats.MikrotiksCount) / float64(stats.MaxMikrotiks) * 100
	}
	if stats.MaxNetworkUsers > 0 {
		stats.NetworkUsersUsagePercent = float64(stats.NetworkUsersCount) / float64(stats.MaxNetworkUsers) * 100
	}
	if stats.MaxStaffUsers > 0 {
		stats.StaffUsersUsagePercent = float64(stats.StaffUsersCount) / float64(stats.MaxStaffUsers) * 100
	}

	return stats, nil
}

// CheckLimits validates if tenant can add more resources
func (a *tenantAdapter) CheckLimits(ctx context.Context, tenantID uuid.UUID, resourceType string) (bool, error) {
	stats, err := a.GetStats(ctx, tenantID)
	if err != nil {
		return false, err
	}

	switch resourceType {
	case "mikrotik":
		return stats.MikrotiksCount < stats.MaxMikrotiks, nil
	case "network_user":
		return stats.NetworkUsersCount < stats.MaxNetworkUsers, nil
	case "staff_user":
		return stats.StaffUsersCount < stats.MaxStaffUsers, nil
	default:
		return false, fmt.Errorf("invalid resource type: %s", resourceType)
	}
}

// IsSubdomainAvailable checks if subdomain is available
func (a *tenantAdapter) IsSubdomainAvailable(ctx context.Context, subdomain string, excludeTenantID *uuid.UUID) (bool, error) {
	var count int64
	query := a.db.WithContext(ctx).
		Model(&model.Tenant{}).
		Where("subdomain = ?", subdomain)

	if excludeTenantID != nil {
		query = query.Where("id != ?", *excludeTenantID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check subdomain availability: %w", err)
	}

	return count == 0, nil
}
