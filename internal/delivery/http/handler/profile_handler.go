package handler

import (
	"log"
	"mikrobill/internal/entity"
	"mikrobill/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProfileHandler handles HTTP requests for profile management
type ProfileHandler struct {
	service *usecase.ProfileService
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(service *usecase.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		service: service,
	}
}

// CreateProfileRequest represents the request body for creating a profile
type CreateProfileRequest struct {
	MikrotikID           string   `json:"mikrotik_id" binding:"required"`
	Name                 string   `json:"name" binding:"required"`
	ProfileType          string   `json:"profile_type" binding:"required"` // pppoe, hotspot, static_ip
	RateLimitUp          *string  `json:"rate_limit_up"`
	RateLimitDown        *string  `json:"rate_limit_down"`
	IdleTimeout          *string  `json:"idle_timeout"`
	SessionTimeout       *string  `json:"session_timeout"`
	KeepaliveTimeout     *string  `json:"keepalive_timeout"`
	OnlyOne              bool     `json:"only_one"`
	StatusAuthentication bool     `json:"status_authentication"`
	DNSServer            *string  `json:"dns_server"`
	Price                *float64 `json:"price"`
	SyncWithMikrotik     bool     `json:"sync_with_mikrotik"`

	// PPPoE specific fields
	PPPoE *PPPoEDetailsRequest `json:"pppoe_details,omitempty"`
}

// PPPoEDetailsRequest represents PPPoE-specific fields
type PPPoEDetailsRequest struct {
	LocalAddress   string  `json:"local_address" binding:"required"`
	RemoteAddress  *string `json:"remote_address"`
	AddressPool    string  `json:"address_pool" binding:"required"`
	MTU            string  `json:"mtu"`
	MRU            string  `json:"mru"`
	ServiceName    *string `json:"service_name"`
	MaxMTU         *string `json:"max_mtu"`
	MaxMRU         *string `json:"max_mru"`
	UseMPLS        bool    `json:"use_mpls"`
	UseCompression bool    `json:"use_compression"`
	UseEncryption  bool    `json:"use_encryption"`
}

// CreateProfile handles profile creation
// POST /api/profiles
func (h *ProfileHandler) CreateProfile(c *gin.Context) {
	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ProfileHandler] CreateProfile - Validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Validate profile type
	if req.ProfileType == "pppoe" && req.PPPoE == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "pppoe_details required for pppoe profile",
		})
		return
	}

	// Create profile domain model
	profile := &entity.MikrotikProfile{
		ID:                   uuid.New().String(),
		MikrotikID:           req.MikrotikID,
		Name:                 req.Name,
		ProfileType:          req.ProfileType,
		RateLimitUp:          req.RateLimitUp,
		RateLimitDown:        req.RateLimitDown,
		IdleTimeout:          req.IdleTimeout,
		SessionTimeout:       req.SessionTimeout,
		KeepaliveTimeout:     req.KeepaliveTimeout,
		OnlyOne:              req.OnlyOne,
		StatusAuthentication: req.StatusAuthentication,
		DNSServer:            req.DNSServer,
		Price:                req.Price,
		SyncWithMikrotik:     req.SyncWithMikrotik,
		IsActive:             true,
	}

	// Create PPPoE details if provided
	var pppoeDetails *entity.MikrotikProfilePPPoE
	if req.PPPoE != nil {
		mtu := req.PPPoE.MTU
		if mtu == "" {
			mtu = "1480"
		}
		mru := req.PPPoE.MRU
		if mru == "" {
			mru = "1480"
		}

		pppoeDetails = &entity.MikrotikProfilePPPoE{
			ProfileID:      profile.ID,
			LocalAddress:   req.PPPoE.LocalAddress,
			RemoteAddress:  req.PPPoE.RemoteAddress,
			AddressPool:    req.PPPoE.AddressPool,
			MTU:            mtu,
			MRU:            mru,
			ServiceName:    req.PPPoE.ServiceName,
			MaxMTU:         req.PPPoE.MaxMTU,
			MaxMRU:         req.PPPoE.MaxMRU,
			UseMPLS:        req.PPPoE.UseMPLS,
			UseCompression: req.PPPoE.UseCompression,
			UseEncryption:  req.PPPoE.UseEncryption,
		}
	}

	// Create profile with sync
	if err := h.service.CreateProfileWithSync(profile, pppoeDetails); err != nil {
		log.Printf("[ProfileHandler] CreateProfile - Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   profile,
	})
}

// UpdateProfile handles profile updates
// PUT /api/profiles/:id
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	id := c.Param("id")

	var req CreateProfileRequest // Reuse the same struct
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ProfileHandler] UpdateProfile - Validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Validate profile type
	if req.ProfileType == "pppoe" && req.PPPoE == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "pppoe_details required for pppoe profile",
		})
		return
	}

	// Create profile domain model
	profile := &entity.MikrotikProfile{
		ID:                   id,
		MikrotikID:           req.MikrotikID,
		Name:                 req.Name,
		ProfileType:          req.ProfileType,
		RateLimitUp:          req.RateLimitUp,
		RateLimitDown:        req.RateLimitDown,
		IdleTimeout:          req.IdleTimeout,
		SessionTimeout:       req.SessionTimeout,
		KeepaliveTimeout:     req.KeepaliveTimeout,
		OnlyOne:              req.OnlyOne,
		StatusAuthentication: req.StatusAuthentication,
		DNSServer:            req.DNSServer,
		Price:                req.Price,
		SyncWithMikrotik:     req.SyncWithMikrotik,
		IsActive:             true,
	}

	// Create PPPoE details if provided
	var pppoeDetails *entity.MikrotikProfilePPPoE
	if req.PPPoE != nil {
		mtu := req.PPPoE.MTU
		if mtu == "" {
			mtu = "1480"
		}
		mru := req.PPPoE.MRU
		if mru == "" {
			mru = "1480"
		}

		pppoeDetails = &entity.MikrotikProfilePPPoE{
			ProfileID:      id,
			LocalAddress:   req.PPPoE.LocalAddress,
			RemoteAddress:  req.PPPoE.RemoteAddress,
			AddressPool:    req.PPPoE.AddressPool,
			MTU:            mtu,
			MRU:            mru,
			ServiceName:    req.PPPoE.ServiceName,
			MaxMTU:         req.PPPoE.MaxMTU,
			MaxMRU:         req.PPPoE.MaxMRU,
			UseMPLS:        req.PPPoE.UseMPLS,
			UseCompression: req.PPPoE.UseCompression,
			UseEncryption:  req.PPPoE.UseEncryption,
		}
	}

	// Update profile with sync
	if err := h.service.UpdateProfileWithSync(profile, pppoeDetails); err != nil {
		log.Printf("[ProfileHandler] UpdateProfile - Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile updated successfully",
	})
}

// DeleteProfile handles profile deletion
// DELETE /api/profiles/:id
func (h *ProfileHandler) DeleteProfile(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteProfileWithSync(id); err != nil {
		log.Printf("[ProfileHandler] DeleteProfile - Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile deleted successfully",
	})
}

// GetProfile handles retrieving a single profile
// GET /api/profiles/:id
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")

	profile, err := h.service.GetProfile(id)
	if err != nil {
		log.Printf("[ProfileHandler] GetProfile - Service error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   profile,
	})
}

// ListProfiles handles listing profiles with pagination
// GET /api/profiles
func (h *ProfileHandler) ListProfiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var mikrotikID *string
	if mtID := c.Query("mikrotik_id"); mtID != "" {
		mikrotikID = &mtID
	}

	profiles, total, err := h.service.ListProfiles(mikrotikID, page, limit)
	if err != nil {
		log.Printf("[ProfileHandler] ListProfiles - Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   profiles,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// SyncProfile handles syncing a single profile to MikroTik
// POST /api/profiles/:id/sync
func (h *ProfileHandler) SyncProfile(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.SyncProfileToMikrotik(id); err != nil {
		log.Printf("[ProfileHandler] SyncProfile - Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile synced to MikroTik successfully",
	})
}

// SyncAllProfiles handles syncing all profiles from MikroTik
// POST /api/profiles/sync-all/:mikrotik_id
func (h *ProfileHandler) SyncAllProfiles(c *gin.Context) {
	mikrotikID := c.Param("mikrotik_id")

	if err := h.service.SyncAllProfiles(mikrotikID); err != nil {
		log.Printf("[ProfileHandler] SyncAllProfiles - Service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "All profiles synced from MikroTik successfully",
	})
}
