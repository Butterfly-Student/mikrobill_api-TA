// internal/handler/role_handler.go
package handler

// import (
// 	"net/http"
// 	"strconv"

// 	request_dto "Mikrotik-Layer/internal/dto/request"
// 	response_dto "Mikrotik-Layer/internal/dto/response"
// 	"Mikrotik-Layer/internal/service"
// 	"Mikrotik-Layer/internal/utils"
// 	pkg_logger "Mikrotik-Layer/pkg/logger"

// 	"github.com/gin-gonic/gin"
// 	"go.uber.org/zap"
// )

// // RoleHandler handles all role management operations
// type RoleHandler struct {
// 	roleService service.RoleService
// }

// // NewRoleHandler creates a new instance of RoleHandler
// func NewRoleHandler(roleService service.RoleService) *RoleHandler {
// 	return &RoleHandler{
// 		roleService: roleService,
// 	}
// }

// // Create creates a new role
// func (h *RoleHandler) Create(c *gin.Context) {
// 	var req request_dto.CreateRoleRequest

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		pkg_logger.Warn("Invalid create role request", zap.Error(err))
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
// 		return
// 	}

// 	role, err := h.roleService.Create(c.Request.Context(), req)
// 	if err != nil {
// 		pkg_logger.Error("Failed to create role",
// 			zap.Error(err),
// 			zap.String("name", req.Name),
// 		)
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to create role", err)
// 		return
// 	}

// 	pkg_logger.Info("Role created successfully",
// 		zap.Int64("role_id", role.ID),
// 		zap.String("name", role.Name),
// 	)

// 	utils.SuccessResponse(c, http.StatusCreated, "Role created successfully", gin.H{
// 		"id":           role.ID,
// 		"name":         role.Name,
// 		"display_name": role.DisplayName,
// 		"description":  role.Description,
// 	})
// }

// // GetByID gets a role by ID
// func (h *RoleHandler) GetByID(c *gin.Context) {
// 	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err)
// 		return
// 	}

// 	role, err := h.roleService.GetByID(c.Request.Context(), id)
// 	if err != nil {
// 		pkg_logger.Error("Failed to get role",
// 			zap.Error(err),
// 			zap.Int64("role_id", id),
// 		)
// 		utils.ErrorResponse(c, http.StatusNotFound, "Role not found", err)
// 		return
// 	}

// 	utils.SuccessResponse(c, http.StatusOK, "Role retrieved successfully", role)
// }

// // List lists all roles with pagination
// func (h *RoleHandler) List(c *gin.Context) {
// 	var req request_dto.PaginationRequest

// 	if err := c.ShouldBindQuery(&req); err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
// 		return
// 	}

// 	// Set defaults
// 	if req.Page < 1 {
// 		req.Page = 1
// 	}
// 	if req.PageSize < 1 {
// 		req.PageSize = 10
// 	}

// 	result, err := h.roleService.List(c.Request.Context(), req)
// 	if err != nil {
// 		pkg_logger.Error("Failed to list roles", zap.Error(err))
// 		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list roles", err)
// 		return
// 	}

// 	utils.SuccessResponse(c, http.StatusOK, "Roles retrieved successfully", result)
// }

// // Update updates a role
// func (h *RoleHandler) Update(c *gin.Context) {
// 	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err)
// 		return
// 	}

// 	var req request_dto.UpdateRoleRequest

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
// 		return
// 	}

// 	if err := h.roleService.Update(c.Request.Context(), id, req); err != nil {
// 		pkg_logger.Error("Failed to update role",
// 			zap.Error(err),
// 			zap.Int64("role_id", id),
// 		)
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update role", err)
// 		return
// 	}

// 	pkg_logger.Info("Role updated successfully",
// 		zap.Int64("role_id", id),
// 	)

// 	utils.SuccessResponse(c, http.StatusOK, "Role updated successfully", nil)
// }

// // Delete deletes a role
// func (h *RoleHandler) Delete(c *gin.Context) {
// 	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err)
// 		return
// 	}

// 	if err := h.roleService.Delete(c.Request.Context(), id); err != nil {
// 		pkg_logger.Error("Failed to delete role",
// 			zap.Error(err),
// 			zap.Int64("role_id", id),
// 		)
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to delete role", err)
// 		return
// 	}

// 	pkg_logger.Info("Role deleted successfully",
// 		zap.Int64("role_id", id),
// 	)

// 	utils.SuccessResponse(c, http.StatusOK, "Role deleted successfully", nil)
// }

// // UpdatePermissions updates role permissions
// func (h *RoleHandler) UpdatePermissions(c *gin.Context) {
// 	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err)
// 		return
// 	}

// 	var req struct {
// 		Permissions []response_dto.Permission `json:"permissions" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
// 		return
// 	}

// 	if err := h.roleService.UpdatePermissions(c.Request.Context(), id, req.Permissions); err != nil {
// 		pkg_logger.Error("Failed to update role permissions",
// 			zap.Error(err),
// 			zap.Int64("role_id", id),
// 		)
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update permissions", err)
// 		return
// 	}

// 	pkg_logger.Info("Role permissions updated successfully",
// 		zap.Int64("role_id", id),
// 		zap.Int("permission_count", len(req.Permissions)),
// 	)

// 	utils.SuccessResponse(c, http.StatusOK, "Permissions updated successfully", nil)
// }

// // GetRoleUsers gets all users assigned to a role
// func (h *RoleHandler) GetRoleUsers(c *gin.Context) {
// 	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid role ID", err)
// 		return
// 	}

// 	users, err := h.roleService.GetRoleUsers(c.Request.Context(), id)
// 	if err != nil {
// 		pkg_logger.Error("Failed to get role users",
// 			zap.Error(err),
// 			zap.Int64("role_id", id),
// 		)
// 		utils.ErrorResponse(c, http.StatusNotFound, "Failed to get role users", err)
// 		return
// 	}

// 	utils.SuccessResponse(c, http.StatusOK, "Role users retrieved successfully", users)
// }