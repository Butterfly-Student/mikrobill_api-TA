package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"prabogo/internal/domain"
	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
)

type profileAdapter struct {
	domainRegistry domain.Domain
}

func NewProfileAdapter(domainRegistry domain.Domain) inbound_port.ProfilePort {
	return &profileAdapter{
		domainRegistry: domainRegistry,
	}
}

func (a *profileAdapter) CreateProfile(ctx any) error {
	c := ctx.(*gin.Context)

	var input model.ProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to bind request")
	}

	profile, err := a.domainRegistry.Profile().CreateProfile(c, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create profile",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to create profile")
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "profile created successfully",
		"data":    profile,
	})

	return nil
}

func (a *profileAdapter) GetProfile(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	profile, err := a.domainRegistry.Profile().GetProfile(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get profile",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to get profile")
	}

	c.JSON(http.StatusOK, gin.H{
		"data": profile,
	})

	return nil
}

func (a *profileAdapter) ListProfiles(ctx any) error {
	c := ctx.(*gin.Context)

	profiles, err := a.domainRegistry.Profile().ListProfiles(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list profiles",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to list profiles")
	}

	c.JSON(http.StatusOK, gin.H{
		"data": profiles,
	})

	return nil
}

func (a *profileAdapter) UpdateProfile(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	var input model.ProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to bind request")
	}

	profile, err := a.domainRegistry.Profile().UpdateProfile(c, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update profile",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to update profile")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "profile updated successfully",
		"data":    profile,
	})

	return nil
}

func (a *profileAdapter) DeleteProfile(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	err := a.domainRegistry.Profile().DeleteProfile(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete profile",
			"details": err.Error(),
		})
		return stacktrace.Propagate(err, "failed to delete profile")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "profile deleted successfully",
	})

	return nil
}
