package gin_inbound_adapter

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"prabogo/internal/model"
	contextutil "prabogo/utils/context"
	"prabogo/utils/logger"
)

// TenantContext middleware resolves tenant from authenticated context
// This middleware must be called AFTER authentication middleware
func (h *middlewareAdapter) TenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		l := logger.GetLogger()

		// Get authenticated user from context (set by previous auth middleware)
		user, err := h.getUserFromGinContext(c)
		if err != nil {
			l.Error("Failed to get user from context", zap.Error(err))
			SendAbort(c, http.StatusUnauthorized, "User not authenticated")
			return
		}

		// Resolve tenant ID based on user type and request headers
		tenantID, err := h.resolveTenantID(c, user)
		if err != nil {
			l.Error("Failed to resolve tenant ID",
				zap.String("user_id", user.ID.String()),
				zap.Error(err))
			SendAbort(c, http.StatusForbidden, err.Error())
			return
		}

		// Validate user has access to the resolved tenant
		if !user.IsSuperadmin {
			hasAccess, err := h.validateTenantAccess(user.ID, tenantID)
			if err != nil {
				l.Error("Failed to validate tenant access",
					zap.String("user_id", user.ID.String()),
					zap.String("tenant_id", tenantID.String()),
					zap.Error(err))
				SendAbort(c, http.StatusInternalServerError, "Failed to validate tenant access")
				return
			}

			if !hasAccess {
				l.Warn("User attempted to access unauthorized tenant",
					zap.String("user_id", user.ID.String()),
					zap.String("tenant_id", tenantID.String()))
				SendAbort(c, http.StatusForbidden, "Access denied to requested tenant")
				return
			}
		}

		// Set tenant context in Gin context
		c.Set("tenant_id", tenantID)
		c.Set("user", user)
		c.Set("is_superadmin", user.IsSuperadmin)

		// Also set in request context for repository layer
		ctx := c.Request.Context()
		ctx = contextutil.WithTenantContext(ctx, tenantID, user, user.IsSuperadmin)
		c.Request = c.Request.WithContext(ctx)

		// Log tenant context resolution
		l.Debug("Tenant context resolved",
			zap.String("user_id", user.ID.String()),
			zap.String("tenant_id", tenantID.String()),
			zap.Bool("is_superadmin", user.IsSuperadmin))

		c.Next()
	}
}

// getUserFromGinContext retrieves user from Gin context
// User should be set by previous authentication middleware
func (h *middlewareAdapter) getUserFromGinContext(c *gin.Context) (*model.User, error) {
	userValue, exists := c.Get("user")
	if !exists {
		return nil, fmt.Errorf("user not found in context")
	}

	user, ok := userValue.(*model.User)
	if !ok {
		return nil, fmt.Errorf("invalid user type in context")
	}

	return user, nil
}

// resolveTenantID determines the tenant ID based on user and request headers
func (h *middlewareAdapter) resolveTenantID(c *gin.Context, user *model.User) (uuid.UUID, error) {
	// Strategy 1: Check X-Tenant-ID header (for super admin tenant selection or multi-tenant users)
	tenantIDHeader := c.GetHeader("X-Tenant-ID")
	if tenantIDHeader != "" {
		tenantID, err := uuid.Parse(tenantIDHeader)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid X-Tenant-ID header format: %w", err)
		}

		// For super admin, header is required but no validation needed
		// For regular users, validation will happen in validateTenantAccess
		return tenantID, nil
	}

	// Strategy 2: For super admin WITHOUT header, deny access (security by design)
	if user.IsSuperadmin {
		return uuid.Nil, fmt.Errorf("super admin must specify X-Tenant-ID header to access tenant resources")
	}

	// Strategy 3: Use user's primary tenant from JWT (tenant_id field)
	if user.TenantID != nil && *user.TenantID != uuid.Nil {
		return *user.TenantID, nil
	}

	// Strategy 4: Query primary tenant from tenant_users table
	primaryTenantID, err := h.getPrimaryTenantForUser(user.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get primary tenant: %w", err)
	}

	if primaryTenantID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("user has no associated tenants")
	}

	return primaryTenantID, nil
}

// validateTenantAccess checks if user has access to the specified tenant
func (h *middlewareAdapter) validateTenantAccess(userID, tenantID uuid.UUID) (bool, error) {
	return h.domain.Database().TenantUser().HasAccess(contextutil.SetUser(contextutil.SetSuperAdmin(contextutil.SetTenantID(context.Background(), tenantID), true), &model.User{ID: userID}), userID, tenantID)
}

// getPrimaryTenantForUser retrieves the primary tenant for a user from tenant_users table
func (h *middlewareAdapter) getPrimaryTenantForUser(userID uuid.UUID) (uuid.UUID, error) {
	return h.domain.Database().TenantUser().GetPrimaryTenant(contextutil.SetUser(context.Background(), &model.User{ID: userID, IsSuperadmin: true}), userID)
}
