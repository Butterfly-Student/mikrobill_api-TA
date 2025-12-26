// internal/handler/user_handler.go
package handler

// import (
// 	"net/http"

// 	request_dto "Mikrotik-Layer/internal/dto/request"
// 	"Mikrotik-Layer/internal/service"
// 	"Mikrotik-Layer/internal/utils"
// 	"Mikrotik-Layer/pkg/logger"

// 	"github.com/gin-gonic/gin"
// 	"go.uber.org/zap"
// )

// type UserHandler struct {
// 	userService service.UserService
// }

// func NewUserHandler(userService service.UserService) *UserHandler {
// 	return &UserHandler{
// 		userService: userService,
// 	}
// }

// func (h *UserHandler) GetByID(c *gin.Context) {
// 	id := c.GetInt64("id")
// 	if id == 0 {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
// 		return
// 	}

// 	user, err := h.userService.GetByID(c.Request.Context(), id)
// 	if err != nil {
// 		utils.ErrorResponse(c, http.StatusNotFound, "User not found", err)
// 		return
// 	}

// 	utils.SuccessResponse(c, http.StatusOK, "User retrieved", user)
// }

// func (h *UserHandler) List(c *gin.Context) {
// 	var req request_dto.PaginationRequest
// 	if err := c.ShouldBindQuery(&req); err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
// 		return
// 	}

// 	result, err := h.userService.List(c.Request.Context(), req)
// 	if err != nil {
// 		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users", err)
// 		return
// 	}

// 	utils.SuccessResponse(c, http.StatusOK, "Users retrieved", result)
// }

// func (h *UserHandler) Update(c *gin.Context) {
// 	id := c.GetInt64("id")
// 	if id == 0 {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
// 		return
// 	}

// 	var req request_dto.UpdateUserRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
// 		return
// 	}

// 	if err := h.userService.Update(c.Request.Context(), id, req); err != nil {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update user", err)
// 		return
// 	}

// 	pkg_logger.Info("User updated", zap.Int64("id", id))
// 	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", nil)
// }

// func (h *UserHandler) Delete(c *gin.Context) {
// 	id := c.GetInt64("id")
// 	if id == 0 {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
// 		return
// 	}

// 	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
// 		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", err)
// 		return
// 	}

// 	pkg_logger.Info("User deleted", zap.Int64("id", id))
// 	utils.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
// }

// func (h *UserHandler) Create(c *gin.Context) {
//     var req request_dto.CreateUserRequest
//     if err := c.ShouldBindJSON(&req); err != nil {
//         utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
//         return
//     }

//     user, err := h.userService.Create(c.Request.Context(), req)
//     if err != nil {
//         utils.ErrorResponse(c, http.StatusBadRequest, "Failed to create user", err)
//         return
//     }

//     pkg_logger.Info("User created", zap.String("email", user.Email))
//     utils.SuccessResponse(c, http.StatusCreated, "User created successfully", user)
// }


// func (h *UserHandler) AssignRoles(c *gin.Context) {
//     userID := c.GetInt64("user_id")
//     if userID == 0 {
//         utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
//         return
//     }

//     var req request_dto.AssignRolesToUserRequest
//     if err := c.ShouldBindJSON(&req); err != nil {
//         utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
//         return
//     }

//     if err := h.userService.AssignRoles(c.Request.Context(), userID, req.RoleIDs); err != nil {
//         utils.ErrorResponse(c, http.StatusBadRequest, "Failed to assign roles", err)
//         return
//     }

//     pkg_logger.Info("Roles assigned to user", zap.Int64("user_id", userID))
//     utils.SuccessResponse(c, http.StatusOK, "Roles assigned successfully", nil)
// }