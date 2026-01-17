package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	outbound_port "MikrOps/internal/port/outbound"
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

func (d *profileDomain) CreateProfile(ctx context.Context, input model.CreateProfileRequest) (*model.MikrotikProfile, error) {
	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	mikrotikID, err := uuid.Parse(activeMikrotik.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid mikrotik id")
	}

	// Begin database transaction
	result, err := d.databasePort.DoInTransaction(ctx, func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Insert to mikrotik_profiles
		profile, err := txDB.Profile().CreateProfile(ctx, input, mikrotikID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create profile")
		}

		// 2. Insert to mikrotik_profile_pppoe if PPPoE type
		if input.Type == model.ProfileTypePPPoE {
			profileID, err := uuid.Parse(profile.ID)
			if err != nil {
				return nil, stacktrace.Propagate(err, "invalid profile id")
			}

			err = txDB.Profile().CreateProfilePPPoE(ctx, profileID, input)
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to create profile pppoe")
			}
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

		// Build rate-limit from RateLimit field (format: upload/download)
		if input.RateLimit != nil {
			args["rate-limit"] = *input.RateLimit
		}

		if input.SessionTimeout != nil {
			args["session-timeout"] = fmt.Sprintf("%d", *input.SessionTimeout)
		}

		if input.IdleTimeout != nil {
			args["idle-timeout"] = fmt.Sprintf("%d", *input.IdleTimeout)
		}

		// 5. Call MikroTik API to create profile
		reply, err := client.RunArgs("/ppp/profile/add", args)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create profile in mikrotik")
		}

		// 6. Extract mikrotik object ID from reply
		var mikrotikObjectID string
		if reply.Done != nil && reply.Done.Map != nil {
			if ret, ok := reply.Done.Map["ret"]; ok {
				mikrotikObjectID = ret
			} else if after, ok := reply.Done.Map["after"]; ok {
				mikrotikObjectID = after
			}
		}
		if mikrotikObjectID == "" {
			return nil, fmt.Errorf("failed to get mikrotik object id from response")
		}

		// 7. Update mikrotik_object_id in database
		profileID, err := uuid.Parse(profile.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid profile id")
		}

		err = txDB.Profile().UpdateMikrotikObjectID(ctx, profileID, mikrotikObjectID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update mikrotik object id")
		}

		// 8. Get complete profile with PPPoE settings
		profileWithDetails, err := txDB.Profile().GetByID(ctx, profileID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get created profile")
		}

		return profileWithDetails, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.MikrotikProfile), nil
}

func (d *profileDomain) GetProfile(ctx context.Context, id string) (*model.MikrotikProfile, error) {
	profileID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid profile id")
	}

	profile, err := d.databasePort.Profile().GetByID(ctx, profileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}

	return profile, nil
}

func (d *profileDomain) ListProfiles(ctx context.Context) ([]model.MikrotikProfile, error) {
	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	mikrotikID, err := uuid.Parse(activeMikrotik.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid mikrotik id")
	}

	profiles, err := d.databasePort.Profile().List(ctx, mikrotikID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list profiles")
	}

	return profiles, nil
}

func (d *profileDomain) UpdateProfile(ctx context.Context, id string, input model.CreateProfileRequest) (*model.MikrotikProfile, error) {
	profileID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid profile id")
	}

	// Get profile to get mikrotik_object_id
	existing, err := d.databasePort.Profile().GetByID(ctx, profileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get existing profile")
	}

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	result, err := d.databasePort.DoInTransaction(ctx, func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Update database
		err := txDB.Profile().Update(ctx, profileID, input)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update profile in database")
		}

		// 2. Update MikroTik if has object ID
		if existing.Metadata != nil {
			client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
			}
			defer client.Close()

			// Get mikrotik object ID from metadata or separate field
			// For now assume we have it stored somewhere accessible
			// This would need to be adjusted based on actual schema
			mikrotikObjectID := "" // TODO: Get from metadata or add field to model

			if mikrotikObjectID != "" {
				// Prepare update parameters
				args := make(map[string]string)
				args[".id"] = mikrotikObjectID
				args["name"] = input.Name

				if input.LocalAddress != nil {
					args["local-address"] = *input.LocalAddress
				}
				if input.RemoteAddress != nil {
					args["remote-address"] = *input.RemoteAddress
				}

				if input.RateLimit != nil {
					args["rate-limit"] = *input.RateLimit
				}

				if input.SessionTimeout != nil {
					args["session-timeout"] = fmt.Sprintf("%d", *input.SessionTimeout)
				}
				if input.IdleTimeout != nil {
					args["idle-timeout"] = fmt.Sprintf("%d", *input.IdleTimeout)
				}

				_, err = client.RunArgs("/ppp/profile/set", args)
				if err != nil {
					return nil, stacktrace.Propagate(err, "failed to update profile in mikrotik")
				}
			}
		}

		// 3. Get updated profile
		updated, err := txDB.Profile().GetByID(ctx, profileID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get updated profile")
		}

		return updated, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.MikrotikProfile), nil
}

func (d *profileDomain) DeleteProfile(ctx context.Context, id string) error {
	profileID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid profile id")
	}

	// Get profile to get mikrotik_object_id
	existing, err := d.databasePort.Profile().GetByID(ctx, profileID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get existing profile")
	}

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	_, err = d.databasePort.DoInTransaction(ctx, func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Delete from MikroTik first if has object ID
		if existing.Metadata != nil {
			client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
			}
			defer client.Close()

			mikrotikObjectID := "" // TODO: Get from metadata

			if mikrotikObjectID != "" {
				_, err = client.RunArgs("/ppp/profile/remove", map[string]string{
					".id": mikrotikObjectID,
				})
				if err != nil {
					return nil, stacktrace.Propagate(err, "failed to delete profile from mikrotik")
				}
			}
		}

		// 2. Delete from database
		err = txDB.Profile().Delete(ctx, profileID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to delete profile from database")
		}

		return nil, nil
	})

	return err
}

