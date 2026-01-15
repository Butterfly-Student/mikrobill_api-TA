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
		SendError(c, http.StatusInternalServerError, "Failed to create profile")
		return stacktrace.Propagate(err, "failed to create profile")
	}

	SendResponse(c, http.StatusCreated, profile, nil)

	return nil
}

func (a *profileAdapter) GetProfile(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	profile, err := a.domainRegistry.Profile().GetProfile(c, id)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to get profile")
		return stacktrace.Propagate(err, "failed to get profile")
	}

	SendResponse(c, http.StatusOK, profile, nil)

	return nil
}

func (a *profileAdapter) ListProfiles(ctx any) error {
	c := ctx.(*gin.Context)

	profiles, err := a.domainRegistry.Profile().ListProfiles(c)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to list profiles")
		return stacktrace.Propagate(err, "failed to list profiles")
	}

	SendResponse(c, http.StatusOK, profiles, &model.Metadata{
		Total: int64(len(profiles)),
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
		SendError(c, http.StatusInternalServerError, "Failed to update profile")
		return stacktrace.Propagate(err, "failed to update profile")
	}

	SendResponse(c, http.StatusOK, profile, nil)

	return nil
}

func (a *profileAdapter) DeleteProfile(ctx any) error {
	c := ctx.(*gin.Context)
	id := c.Param("id")

	err := a.domainRegistry.Profile().DeleteProfile(c, id)
	if err != nil {
		SendError(c, http.StatusInternalServerError, "Failed to delete profile")
		return stacktrace.Propagate(err, "failed to delete profile")
	}

	SendResponse(c, http.StatusOK, nil, nil)

	return nil
}
