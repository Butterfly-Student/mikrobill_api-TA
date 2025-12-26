// internal/middleware/casbin.go
package middleware

import (
	pkg_logger "mikrobill/pkg/logger"
	"mikrobill/pkg/utils"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CasbinMiddleware enforces RBAC using Casbin
// Expects user_role and user_roles from AuthMiddleware in context
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user info from context (set by AuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, 401, "User not authenticated", utils.ErrUnauthorized)
			c.Abort()
			return
		}

		// Get single role (string)
		role, exists := c.Get("user_role")
		if !exists || role == nil {
			utils.ErrorResponse(c, 403, "User role not found", utils.ErrForbidden)
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok || userRole == "" {
			pkg_logger.Warn("Invalid user role format",
				zap.Int64("user_id", userID.(int64)),
			)
			utils.ErrorResponse(c, 403, "Invalid user role", utils.ErrForbidden)
			c.Abort()
			return
		}

		// Resource = endpoint path, Action = HTTP method
		resource := c.FullPath() // e.g., "/api/v1/users/:id"
		action := c.Request.Method // e.g., "GET", "POST"

		// Skip authentication check for public endpoints
		if resource == "" {
			pkg_logger.Debug("Skipping Casbin check for empty resource path")
			c.Next()
			return
		}

		// Enforce permission check
		allowed, err := enforcer.Enforce(userRole, resource, action)
		if err != nil {
			pkg_logger.Error("Casbin enforcement error",
				zap.Error(err),
				zap.String("role", userRole),
				zap.String("resource", resource),
				zap.String("action", action),
			)
			utils.ErrorResponse(c, 500, "Permission check failed", err)
			c.Abort()
			return
		}

		pkg_logger.Debug("Casbin enforcement check",
			zap.Int64("user_id", userID.(int64)),
			zap.String("role", userRole),
			zap.String("resource", resource),
			zap.String("action", action),
			zap.Bool("allowed", allowed),
		)

		if !allowed {
			pkg_logger.Warn("Permission denied",
				zap.Int64("user_id", userID.(int64)),
				zap.String("role", userRole),
				zap.String("resource", resource),
				zap.String("action", action),
			)
			utils.ErrorResponse(c, 403, "Permission denied for this resource", utils.ErrPermissionDenied)
			c.Abort()
			return
		}

		pkg_logger.Info("Permission granted",
			zap.Int64("user_id", userID.(int64)),
			zap.String("role", userRole),
			zap.String("resource", resource),
			zap.String("action", action),
		)

		c.Next()
	}
}

// LoadCasbinPolicies loads policies from database via adapter
func LoadCasbinPolicies(enforcer *casbin.Enforcer) error {
	pkg_logger.Info("Loading Casbin policies from database")
	
	// Adapter akan otomatis load dari casbin_rule table
	if err := enforcer.LoadPolicy(); err != nil {
		pkg_logger.Error("Failed to load Casbin policies", zap.Error(err))
		return err
	}

	// Log loaded policies for debugging
	policies, _ := enforcer.GetPolicy()
	pkg_logger.Info("Casbin policies loaded",
		zap.Int("policy_count", len(policies)),
	)

	return nil
}

// SyncRolePermissions syncs permissions from roles table to Casbin policies
// This should be called when role permissions are updated
func SyncRolePermissions(enforcer *casbin.Enforcer, roleName string, permissions []struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
	Path     string   `json:"path"`
}) error {
	pkg_logger.Info("Syncing role permissions to Casbin",
		zap.String("role", roleName),
	)

	// Remove old policies for this role
	_, err := enforcer.RemoveFilteredPolicy(0, roleName)
	if err != nil {
		pkg_logger.Error("Failed to remove old policies", zap.Error(err))
		return err
	}

	// Add new policies
	policyCount := 0
	for _, perm := range permissions {
		for _, action := range perm.Actions {
			method := actionToMethod(action)
			_, err := enforcer.AddPolicy(roleName, perm.Path, method)
			if err != nil {
				pkg_logger.Warn("Failed to add policy",
					zap.Error(err),
					zap.String("role", roleName),
					zap.String("path", perm.Path),
					zap.String("method", method),
				)
			} else {
				policyCount++
			}
		}
	}

	// Save to database
	if err := enforcer.SavePolicy(); err != nil {
		pkg_logger.Error("Failed to save policies", zap.Error(err))
		return err
	}

	pkg_logger.Info("Role permissions synced",
		zap.String("role", roleName),
		zap.Int("policies_added", policyCount),
	)

	return nil
}

// AddPolicy adds a new policy and saves to database
func AddPolicy(enforcer *casbin.Enforcer, role, resource, action string) error {
	added, err := enforcer.AddPolicy(role, resource, action)
	if err != nil {
		return err
	}

	if !added {
		pkg_logger.Warn("Policy already exists",
			zap.String("role", role),
			zap.String("resource", resource),
			zap.String("action", action),
		)
		return nil
	}

	// SavePolicy persists to database via adapter
	if err := enforcer.SavePolicy(); err != nil {
		pkg_logger.Error("Failed to save policy", zap.Error(err))
		return err
	}

	pkg_logger.Info("Policy added and saved",
		zap.String("role", role),
		zap.String("resource", resource),
		zap.String("action", action),
	)

	return nil
}

// RemovePolicy removes a policy and saves to database
func RemovePolicy(enforcer *casbin.Enforcer, role, resource, action string) error {
	removed, err := enforcer.RemovePolicy(role, resource, action)
	if err != nil {
		return err
	}

	if !removed {
		pkg_logger.Warn("Policy does not exist",
			zap.String("role", role),
			zap.String("resource", resource),
			zap.String("action", action),
		)
		return nil
	}

	// SavePolicy persists to database via adapter
	if err := enforcer.SavePolicy(); err != nil {
		pkg_logger.Error("Failed to save policy", zap.Error(err))
		return err
	}

	pkg_logger.Info("Policy removed and saved",
		zap.String("role", role),
		zap.String("resource", resource),
		zap.String("action", action),
	)

	return nil
}

// actionToMethod converts action name to HTTP method
func actionToMethod(action string) string {
	switch action {
	case "read":
		return "GET"
	case "create":
		return "POST"
	case "update":
		return "PUT"
	case "delete":
		return "DELETE"
	case "sync", "test", "export":
		return "POST"
	default:
		return "GET"
	}
}