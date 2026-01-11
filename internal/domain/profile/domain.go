package profile

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	outbound_port "prabogo/internal/port/outbound"
)

type profileDomain struct {
	databasePort          outbound_port.DatabasePort
	mikrotikClientFactory outbound_port.MikrotikClientFactory
}

func NewProfileDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
) inbound_port.ProfileDomain {
	return &profileDomain{
		databasePort:          databasePort,
		mikrotikClientFactory: mikrotikClientFactory,
	}
}

func (d *profileDomain) CreateProfile(ctx any, input model.ProfileInput) (*model.ProfileWithPPPoE, error) {
	// Validate and prepare input
	model.PrepareProfileInput(&input)

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	// Begin database transaction
	result, err := d.databasePort.DoInTransaction(func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Insert to mikrotik_profiles
		profile, err := txDB.Profile().CreateProfile(input, activeMikrotik.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create profile")
		}

		// 2. Insert to mikrotik_profile_pppoe
		err = txDB.Profile().CreateProfilePPPoE(profile.ID, input)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create profile pppoe")
		}

		// 3. Create MikroTik client
		client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
		}
		defer client.Close()

		// 4. Prepare MikroTik API parameters
		args := make(map[string]string)
		args["name"] = input.Name

		if input.LocalAddress != nil {
			args["local-address"] = *input.LocalAddress
		}
		if input.RemoteAddress != nil {
			args["remote-address"] = *input.RemoteAddress
		}

		// Build rate-limit (format: upload/download)
		if input.RateLimitUpKbps != nil && input.RateLimitDownKbps != nil {
			rateLimit := fmt.Sprintf("%dk/%dk", *input.RateLimitUpKbps, *input.RateLimitDownKbps)
			args["rate-limit"] = rateLimit
		}

		if input.OnlyOne != nil && *input.OnlyOne {
			args["only-one"] = "yes"
		} else {
			args["only-one"] = "no"
		}

		if input.SessionTimeoutSeconds != nil {
			args["session-timeout"] = fmt.Sprintf("%d", *input.SessionTimeoutSeconds)
		}

		if input.IdleTimeoutSeconds != nil {
			args["idle-timeout"] = fmt.Sprintf("%d", *input.IdleTimeoutSeconds)
		}

		if input.KeepaliveTimeoutSeconds != nil {
			args["keepalive-timeout"] = fmt.Sprintf("%d", *input.KeepaliveTimeoutSeconds)
		}

		if input.DNSServer != nil {
			args["dns-server"] = *input.DNSServer
		}

		// 5. Call MikroTik API to create profile
		reply, err := client.RunArgs("/ppp/profile/add", args)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create profile in mikrotik")
		}

		// 6. Extract mikrotik object ID from reply
		// For add commands, RouterOS returns the ID in the Done response
		var mikrotikObjectID string
		if reply.Done != nil && reply.Done.Map != nil {
			// Try "ret" first (newer API)
			if ret, ok := reply.Done.Map["ret"]; ok {
				mikrotikObjectID = ret
			} else if after, ok := reply.Done.Map["after"]; ok {
				// Fallback to "after" (older API)
				mikrotikObjectID = after
			}
		}
		if mikrotikObjectID == "" {
			return nil, fmt.Errorf("failed to get mikrotik object id from response")
		}

		// 7. Update mikrotik_object_id in database
		err = txDB.Profile().UpdateMikrotikObjectID(profile.ID, mikrotikObjectID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update mikrotik object id")
		}

		// 8. Get complete profile with PPPoE settings
		profileWithPPPoE, err := txDB.Profile().GetByID(profile.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get created profile")
		}

		return profileWithPPPoE, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.ProfileWithPPPoE), nil
}

func (d *profileDomain) GetProfile(ctx any, id string) (*model.ProfileWithPPPoE, error) {
	profileID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid profile id")
	}

	profile, err := d.databasePort.Profile().GetByID(profileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}

	return profile, nil
}

func (d *profileDomain) ListProfiles(ctx any) ([]model.ProfileWithPPPoE, error) {
	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	profiles, err := d.databasePort.Profile().List(activeMikrotik.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list profiles")
	}

	return profiles, nil
}

func (d *profileDomain) UpdateProfile(ctx any, id string, input model.ProfileInput) (*model.ProfileWithPPPoE, error) {
	profileID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid profile id")
	}

	model.PrepareProfileInput(&input)

	// Get profile to get mikrotik_object_id
	existing, err := d.databasePort.Profile().GetByID(profileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get existing profile")
	}

	if existing.MikrotikObjectID == "" {
		return nil, fmt.Errorf("profile has no mikrotik object id")
	}

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	result, err := d.databasePort.DoInTransaction(func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Update database
		err := txDB.Profile().Update(profileID, input)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update profile in database")
		}

		// 2. Update MikroTik
		client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
		}
		defer client.Close()

		// Prepare update parameters
		args := make(map[string]string)
		args[".id"] = existing.MikrotikObjectID
		args["name"] = input.Name

		if input.LocalAddress != nil {
			args["local-address"] = *input.LocalAddress
		}
		if input.RemoteAddress != nil {
			args["remote-address"] = *input.RemoteAddress
		}

		if input.RateLimitUpKbps != nil && input.RateLimitDownKbps != nil {
			rateLimit := fmt.Sprintf("%dk/%dk", *input.RateLimitUpKbps, *input.RateLimitDownKbps)
			args["rate-limit"] = rateLimit
		}

		if input.OnlyOne != nil && *input.OnlyOne {
			args["only-one"] = "yes"
		} else {
			args["only-one"] = "no"
		}

		if input.SessionTimeoutSeconds != nil {
			args["session-timeout"] = fmt.Sprintf("%d", *input.SessionTimeoutSeconds)
		}
		if input.IdleTimeoutSeconds != nil {
			args["idle-timeout"] = fmt.Sprintf("%d", *input.IdleTimeoutSeconds)
		}
		if input.KeepaliveTimeoutSeconds != nil {
			args["keepalive-timeout"] = fmt.Sprintf("%d", *input.KeepaliveTimeoutSeconds)
		}
		if input.DNSServer != nil {
			args["dns-server"] = *input.DNSServer
		}

		_, err = client.RunArgs("/ppp/profile/set", args)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update profile in mikrotik")
		}

		// 3. Get updated profile
		updated, err := txDB.Profile().GetByID(profileID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get updated profile")
		}

		return updated, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.ProfileWithPPPoE), nil
}

func (d *profileDomain) DeleteProfile(ctx any, id string) error {
	profileID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid profile id")
	}

	// Get profile to get mikrotik_object_id
	existing, err := d.databasePort.Profile().GetByID(profileID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get existing profile")
	}

	if existing.MikrotikObjectID == "" {
		return fmt.Errorf("profile has no mikrotik object id")
	}

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	_, err = d.databasePort.DoInTransaction(func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Delete from MikroTik first
		client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
		}
		defer client.Close()

		_, err = client.RunArgs("/ppp/profile/remove", map[string]string{
			".id": existing.MikrotikObjectID,
		})
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to delete profile from mikrotik")
		}

		// 2. Delete from database
		err = txDB.Profile().Delete(profileID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to delete profile from database")
		}

		return nil, nil
	})

	return err
}
