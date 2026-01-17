package gin_inbound_adapter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	"MikrOps/utils/activity"
)

type userAdapter struct {
	domain domain.Domain
}

func NewUserAdapter(domain domain.Domain) inbound_port.UserHttpPort {
	return &userAdapter{
		domain: domain,
	}
}

// CreateUser creates a new user (super admin only)
func (a *userAdapter) CreateUser(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_create_user")

	// Get requesting user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "unauthorized",
		})
		return
	}

	// Check if super admin
	role, _ := c.Get("user_role")
	isSuperAdmin := role == string(model.UserRoleSuperAdmin)
	if !isSuperAdmin {
		c.JSON(http.StatusForbidden, model.Response{
			Success: false,
			Error:   "only super admin can create users",
		})
		return
	}

	var request model.CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	user, err := a.domain.User().CreateUser(ctx, request, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Success: true,
		Data:    user,
	})
}

// ListUsers lists users with tenant filtering
func (a *userAdapter) ListUsers(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_list_users")

	userID, _ := c.Get("user_id")
	role, _ := c.Get("user_role")
	isSuperAdmin := role == string(model.UserRoleSuperAdmin)

	// Get pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Get tenant filter (super admin only)
	var tenantID *string
	if isSuperAdmin && c.Query("tenant_id") != "" {
		tid := c.Query("tenant_id")
		tenantID = &tid
	}

	users, total, err := a.domain.User().ListUsers(ctx, tenantID, userID.(string), isSuperAdmin, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data: gin.H{
			"users":  users,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUser retrieves a single user by ID
func (a *userAdapter) GetUser(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_get_user")

	userID, _ := c.Get("user_id")
	role, _ := c.Get("user_role")
	isSuperAdmin := role == string(model.UserRoleSuperAdmin)

	targetUserID := c.Param("id")

	user, err := a.domain.User().GetUserByID(ctx, targetUserID, userID.(string), isSuperAdmin)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    user,
	})
}

// UpdateUser updates a user
func (a *userAdapter) UpdateUser(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_update_user")

	userID, _ := c.Get("user_id")
	role, _ := c.Get("user_role")
	isSuperAdmin := role == string(model.UserRoleSuperAdmin)

	targetUserID := c.Param("id")

	var request model.UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	user, err := a.domain.User().UpdateUser(ctx, targetUserID, request, userID.(string), isSuperAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    user,
	})
}

// DeleteUser soft deletes a user
func (a *userAdapter) DeleteUser(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_delete_user")

	userID, _ := c.Get("user_id")
	role, _ := c.Get("user_role")
	isSuperAdmin := role == string(model.UserRoleSuperAdmin)

	targetUserID := c.Param("id")

	if err := a.domain.User().DeleteUser(ctx, targetUserID, userID.(string), isSuperAdmin); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "user deleted successfully"},
	})
}

// AssignRole assigns a role to a user
func (a *userAdapter) AssignRole(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_assign_role")

	userID, _ := c.Get("user_id")
	role, _ := c.Get("user_role")
	isSuperAdmin := role == string(model.UserRoleSuperAdmin)

	targetUserID := c.Param("id")

	var request struct {
		RoleID string `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if err := a.domain.User().AssignRole(ctx, targetUserID, request.RoleID, userID.(string), isSuperAdmin); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "role assigned successfully"},
	})
}

// AssignToTenant assigns a user to a tenant (super admin only)
func (a *userAdapter) AssignToTenant(i any) {
	c := i.(*gin.Context)
	ctx := activity.NewContext(c.Request.Context(), "http_assign_to_tenant")

	userID, _ := c.Get("user_id")
	role, _ := c.Get("user_role")
	if role != string(model.UserRoleSuperAdmin) {
		c.JSON(http.StatusForbidden, model.Response{
			Success: false,
			Error:   "only super admin can assign users to tenants",
		})
		return
	}

	targetUserID := c.Param("id")

	var request struct {
		TenantID  string  `json:"tenant_id" binding:"required"`
		RoleID    *string `json:"role_id"`
		IsPrimary bool    `json:"is_primary"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if err := a.domain.User().AssignToTenant(ctx, targetUserID, request.TenantID, request.RoleID, request.IsPrimary, userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    gin.H{"message": "user assigned to tenant successfully"},
	})
}
