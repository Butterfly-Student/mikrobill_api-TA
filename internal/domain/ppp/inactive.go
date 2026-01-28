package ppp

import (
	"context"

	"github.com/palantir/stacktrace"

	"MikrOps/internal/model"
)

// --- Inactive PPP Users (Secrets not in Active list) ---

func (d *PPPDomain) MikrotikListInactive(ctx context.Context) ([]model.PPPSecret, error) {
	// Get all secrets
	secrets, err := d.MikrotikListSecrets(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list ppp secrets")
	}

	// Get all active connections
	activeList, err := d.MikrotikListActive(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list active ppp connections")
	}

	// Create map of active usernames for quick lookup
	activeMap := make(map[string]bool)
	for _, active := range activeList {
		activeMap[active.Name] = true
	}

	// Filter secrets that are not active
	var inactive []model.PPPSecret
	for _, secret := range secrets {
		if !activeMap[secret.Name] {
			inactive = append(inactive, secret)
		}
	}

	return inactive, nil
}
