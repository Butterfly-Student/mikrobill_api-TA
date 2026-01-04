package handler

import (
	"log"
	"mikrobill/internal/model"
	"mikrobill/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MikrotikHandler handles HTTP requests for mikrotik management
type MikrotikHandler struct {
	service usecase.MikrotikUseCase
}

// NewMikrotikHandler creates a new mikrotik handler
func NewMikrotikHandler(service usecase.MikrotikUseCase) *MikrotikHandler {
	return &MikrotikHandler{
		service: service,
	}
}

// ListMikrotiks handles listing mikrotiks
// GET /api/mikrotiks
func (h *MikrotikHandler) ListMikrotiks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "100")) // Default larger limit for dropdowns
	search := c.Query("search")

	req := model.PaginationRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	result, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		log.Printf("[MikrotikHandler] ListMikrotiks - Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result.Data,
		"meta": gin.H{
			"page":        result.Page,
			"limit":       result.PageSize,
			"total_items": result.TotalItems,
			"total_pages": result.TotalPages,
		},
	})
}

// CreateMikrotik handles creating a new mikrotik
// POST /api/mikrotiks
func (h *MikrotikHandler) CreateMikrotik(c *gin.Context) {
	var req model.CreateMikrotikRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	mikrotik, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		log.Printf("[MikrotikHandler] CreateMikrotik - Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Mikrotik created successfully",
		"data":    mikrotik,
	})
}
