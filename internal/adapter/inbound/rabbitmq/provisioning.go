package rabbitmq_inbound_adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"MikrOps/internal/domain"
	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	outbound_port "MikrOps/internal/port/outbound"
	"MikrOps/utils/activity"
	contextutil "MikrOps/utils/context"
	"MikrOps/utils/log"
)

type provisioningAdapter struct {
	domain                domain.Domain
	mikrotikClientFactory outbound_port.MikrotikClientFactory
}

func NewProvisioningAdapter(
	domain domain.Domain,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
) inbound_port.ProvisioningMessagePort {
	return &provisioningAdapter{
		domain:                domain,
		mikrotikClientFactory: mikrotikClientFactory,
	}
}

// ProcessProvisioning handles provisioning messages from RabbitMQ
func (h *provisioningAdapter) ProcessProvisioning(a any) bool {
	msg := a.([]byte)
	ctx := activity.NewContext(context.Background(), "rabbitmq_provision_customer")

	var payload model.ProvisioningMessage
	err := json.Unmarshal(msg, &payload)
	if err != nil {
		log.WithContext(ctx).Errorf("provisioning unmarshal error %s: %s", err.Error(), string(msg))
		return true // ACK malformed messages to prevent reprocessing
	}
	ctx = context.WithValue(ctx, activity.Payload, payload)

	// Set tenant context for DB queries
	tenantID, err := uuid.Parse(payload.TenantID)
	if err != nil {
		log.WithContext(ctx).Errorf("invalid tenant_id: %s", err.Error())
		return true // ACK invalid messages
	}
	ctx = contextutil.SetTenantID(ctx, tenantID)

	// Parse IDs
	customerID, err := uuid.Parse(payload.CustomerID)
	if err != nil {
		log.WithContext(ctx).Errorf("invalid customer_id: %s", err.Error())
		return true // ACK invalid messages
	}

	profileID, err := uuid.Parse(payload.ProfileID)
	if err != nil {
		log.WithContext(ctx).Errorf("invalid profile_id: %s", err.Error())
		return true // ACK invalid messages
	}

	// Get customer
	customer, err := h.domain.Database().Customer().GetByID(ctx, customerID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to get customer: %s", err.Error())
		return false // NACK - retry later
	}

	// Get profile
	mikrotikID, err := uuid.Parse(customer.MikrotikID)
	if err != nil {
		log.WithContext(ctx).Errorf("invalid mikrotik_id from customer: %s", err.Error())
		// Update provisioning status to failed
		_ = h.domain.Database().Customer().UpdateProvisioningStatus(ctx, customerID, "failed", fmt.Sprintf("invalid mikrotik_id: %s", err.Error()), nil)
		return true // ACK - cannot recover
	}

	profile, err := h.domain.Database().Profile().GetByMikrotikID(ctx, mikrotikID, profileID)
	if err != nil || profile == nil {
		log.WithContext(ctx).Errorf("failed to get profile: %s", err.Error())
		// Update provisioning status to failed
		_ = h.domain.Database().Customer().UpdateProvisioningStatus(ctx, customerID, "failed", fmt.Sprintf("profile not found: %s", err.Error()), nil)
		return true // ACK - cannot recover
	}

	// Get MikroTik
	mikrotik, err := h.domain.Database().Mikrotik().GetByID(ctx, mikrotikID)
	if err != nil || mikrotik == nil {
		log.WithContext(ctx).Errorf("failed to get mikrotik: %s", err.Error())
		// Update provisioning status to failed
		_ = h.domain.Database().Customer().UpdateProvisioningStatus(ctx, customerID, "failed", fmt.Sprintf("mikrotik not found: %s", err.Error()), nil)
		return true // ACK - cannot recover
	}

	// Create MikroTik client
	client, err := h.mikrotikClientFactory.NewClient(mikrotik)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to create mikrotik client: %s", err.Error())
		// Update provisioning status to failed
		_ = h.domain.Database().Customer().UpdateProvisioningStatus(ctx, customerID, "failed", fmt.Sprintf("failed to create mikrotik client: %s", err.Error()), nil)
		return false // NACK - might be temporary connection issue
	}
	defer client.Close()

	// Provision to MikroTik based on service type
	var mikrotikObjectID string
	if payload.ServiceType == string(model.ServiceTypePPPoE) {
		// Add PPPoE Secret
		args := map[string]string{
			"name":     payload.ServiceUsername,
			"password": payload.ServicePassword,
			"profile":  profile.Name,
			"service":  "pppoe",
		}

		if customer.Name != "" {
			args["comment"] = customer.Name
		}

		// Add local/remote address if configured in profile
		if profile.PPPoEDetails != nil {
			if profile.PPPoEDetails.LocalAddress != nil {
				args["local-address"] = *profile.PPPoEDetails.LocalAddress
			}
			if profile.PPPoEDetails.RemoteAddress != nil {
				args["remote-address"] = *profile.PPPoEDetails.RemoteAddress
			}
		}

		reply, err := client.RunArgs("/ppp/secret/add", args)
		if err != nil {
			log.WithContext(ctx).Errorf("failed to add ppp secret: %s", err.Error())
			// Update provisioning status to failed
			_ = h.domain.Database().Customer().UpdateProvisioningStatus(ctx, customerID, "failed", fmt.Sprintf("mikrotik api error: %s", err.Error()), nil)
			return false // NACK - might be temporary
		}

		// Extract mikrotik object ID
		if reply.Done != nil && reply.Done.Map != nil {
			if ret, ok := reply.Done.Map["ret"]; ok {
				mikrotikObjectID = ret
			} else if after, ok := reply.Done.Map["after"]; ok {
				mikrotikObjectID = after
			}
		}

	} else if payload.ServiceType == string(model.ServiceTypeHotspot) {
		// Add Hotspot User
		args := map[string]string{
			"name":     payload.ServiceUsername,
			"password": payload.ServicePassword,
			"profile":  profile.Name,
		}

		if customer.Name != "" {
			args["comment"] = customer.Name
		}

		reply, err := client.RunArgs("/ip/hotspot/user/add", args)
		if err != nil {
			log.WithContext(ctx).Errorf("failed to add hotspot user: %s", err.Error())
			// Update provisioning status to failed
			_ = h.domain.Database().Customer().UpdateProvisioningStatus(ctx, customerID, "failed", fmt.Sprintf("mikrotik api error: %s", err.Error()), nil)
			return false // NACK - might be temporary
		}

		// Extract mikrotik object ID
		if reply.Done != nil && reply.Done.Map != nil {
			if ret, ok := reply.Done.Map["ret"]; ok {
				mikrotikObjectID = ret
			} else if after, ok := reply.Done.Map["after"]; ok {
				mikrotikObjectID = after
			}
		}
	} else {
		log.WithContext(ctx).Errorf("unsupported service type: %s", payload.ServiceType)
		// Update provisioning status to failed
		_ = h.domain.Database().Customer().UpdateProvisioningStatus(ctx, customerID, "failed", fmt.Sprintf("unsupported service type: %s", payload.ServiceType), nil)
		return true // ACK - cannot recover
	}

	if mikrotikObjectID == "" {
		log.WithContext(ctx).Error("failed to get mikrotik object id from response")
		// Update provisioning status to failed
		_ = h.domain.Database().Customer().UpdateProvisioningStatus(ctx, customerID, "failed", "failed to get mikrotik object id", nil)
		return false // NACK - might be temporary
	}

	// Update customer with mikrotik_object_id and status to active
	err = h.domain.Database().Customer().UpdateMikrotikObjectID(ctx, customerID, mikrotikObjectID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to update mikrotik_object_id: %s", err.Error())
		// Rollback: remove from MikroTik
		if payload.ServiceType == string(model.ServiceTypePPPoE) {
			_, _ = client.RunArgs("/ppp/secret/remove", map[string]string{".id": mikrotikObjectID})
		} else {
			_, _ = client.RunArgs("/ip/hotspot/user/remove", map[string]string{".id": mikrotikObjectID})
		}
		return false // NACK - retry
	}

	// Update provisioning status to active (success)
	now := time.Now()
	err = h.domain.Database().Customer().UpdateProvisioningStatus(ctx, customerID, "active", "", &now)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to update provisioning status to active: %s", err.Error())
		return false // NACK - retry
	}

	ctx = context.WithValue(ctx, activity.Result, map[string]string{
		"customer_id":        payload.CustomerID,
		"service_username":   payload.ServiceUsername,
		"mikrotik_object_id": mikrotikObjectID,
		"status":             "provisioned",
	})

	log.WithContext(ctx).Info("customer provisioning success")
	return true // ACK
}

// ProcessDisconnect handles disconnect messages from RabbitMQ
func (h *provisioningAdapter) ProcessDisconnect(a any) bool {
	msg := a.([]byte)
	ctx := activity.NewContext(context.Background(), "rabbitmq_disconnect_customer")

	var payload model.DisconnectMessage
	err := json.Unmarshal(msg, &payload)
	if err != nil {
		log.WithContext(ctx).Errorf("disconnect unmarshal error %s: %s", err.Error(), string(msg))
		return true // ACK malformed messages
	}
	ctx = context.WithValue(ctx, activity.Payload, payload)

	// Set tenant context for DB queries
	disconnectTenantID, err := uuid.Parse(payload.TenantID)
	if err != nil {
		log.WithContext(ctx).Errorf("invalid tenant_id: %s", err.Error())
		return true // ACK invalid messages
	}
	ctx = contextutil.SetTenantID(ctx, disconnectTenantID)

	// Parse IDs
	customerID, err := uuid.Parse(payload.CustomerID)
	if err != nil {
		log.WithContext(ctx).Errorf("invalid customer_id: %s", err.Error())
		return true // ACK invalid messages
	}

	// Get customer
	customer, err := h.domain.Database().Customer().GetByID(ctx, customerID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to get customer: %s", err.Error())
		return false // NACK - retry
	}

	if customer.MikrotikObjectID == nil || *customer.MikrotikObjectID == "" {
		log.WithContext(ctx).Warn("customer has no mikrotik_object_id, skipping disconnect")
		return true // ACK - nothing to disconnect
	}

	// Get MikroTik
	mikrotikID, err := uuid.Parse(customer.MikrotikID)
	if err != nil {
		log.WithContext(ctx).Errorf("invalid mikrotik_id: %s", err.Error())
		return true // ACK - cannot recover
	}

	mikrotik, err := h.domain.Database().Mikrotik().GetByID(ctx, mikrotikID)
	if err != nil || mikrotik == nil {
		log.WithContext(ctx).Errorf("failed to get mikrotik: %s", err.Error())
		return false // NACK - retry
	}

	// Create MikroTik client
	client, err := h.mikrotikClientFactory.NewClient(mikrotik)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to create mikrotik client: %s", err.Error())
		return false // NACK - temporary
	}
	defer client.Close()

	// Disconnect based on service type
	if payload.ServiceType == string(model.ServiceTypePPPoE) {
		// Find and remove active PPPoE session
		reply, err := client.RunArgs("/ppp/active/print", map[string]string{
			"?name": payload.ServiceUsername,
		})
		if err != nil {
			log.WithContext(ctx).Errorf("failed to find active ppp session: %s", err.Error())
			return false // NACK - retry
		}

		// Remove active sessions
		if reply.Re != nil {
			for _, re := range reply.Re {
				if re.Map != nil {
					if id, ok := re.Map[".id"]; ok {
						_, err = client.RunArgs("/ppp/active/remove", map[string]string{".id": id})
						if err != nil {
							log.WithContext(ctx).Errorf("failed to remove ppp active session: %s", err.Error())
						}
					}
				}
			}
		}

	} else if payload.ServiceType == string(model.ServiceTypeHotspot) {
		// Find and remove active Hotspot session
		reply, err := client.RunArgs("/ip/hotspot/active/print", map[string]string{
			"?user": payload.ServiceUsername,
		})
		if err != nil {
			log.WithContext(ctx).Errorf("failed to find active hotspot session: %s", err.Error())
			return false // NACK - retry
		}

		// Remove active sessions
		if reply.Re != nil {
			for _, re := range reply.Re {
				if re.Map != nil {
					if id, ok := re.Map[".id"]; ok {
						_, err = client.RunArgs("/ip/hotspot/active/remove", map[string]string{".id": id})
						if err != nil {
							log.WithContext(ctx).Errorf("failed to remove hotspot active session: %s", err.Error())
						}
					}
				}
			}
		}
	}

	ctx = context.WithValue(ctx, activity.Result, map[string]string{
		"customer_id":      payload.CustomerID,
		"service_username": payload.ServiceUsername,
		"status":           "disconnected",
	})

	log.WithContext(ctx).Info("customer disconnect success")
	return true // ACK
}
