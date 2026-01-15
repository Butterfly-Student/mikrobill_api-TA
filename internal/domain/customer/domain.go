package customer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	outbound_port "prabogo/internal/port/outbound"
)

type customerDomain struct {
	databasePort          outbound_port.DatabasePort
	mikrotikClientFactory outbound_port.MikrotikClientFactory
	cachePort             outbound_port.CachePort
}

func NewCustomerDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
	cachePort outbound_port.CachePort,
) inbound_port.CustomerDomain {
	return &customerDomain{
		databasePort:          databasePort,
		mikrotikClientFactory: mikrotikClientFactory,
		cachePort:             cachePort,
	}
}

func (d *customerDomain) CreateCustomer(ctx context.Context, input model.CustomerInput) (*model.CustomerWithService, error) { // Validate and prepare input
	model.PrepareCustomerInput(&input)

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	// Validate profile exists
	profile, err := d.databasePort.Profile().GetByMikrotikID(ctx, activeMikrotik.ID, input.ProfileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}
	if profile == nil {
		return nil, fmt.Errorf("profile not found")
	}

	// Check if customer already exists
	existingCustomer, err := d.databasePort.Customer().GetByUsername(ctx, activeMikrotik.ID, input.Username)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing customer")
	}
	if existingCustomer != nil {
		return nil, fmt.Errorf("customer with username '%s' already exists", input.Username)
	}

	// Create MikroTik client
	client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
	}
	defer client.Close()

	// Prepare MikroTik PPP Secret parameters
	// Note: We currently assume PPP Secret for simplicity. Hotspot/Static IP user creation logic would diverge here.
	args := map[string]string{
		"name":     input.Username,
		"password": input.Password,
		"profile":  profile.Name,
		"service":  "pppoe", // Defaulting to pppoe, or use string(input.ServiceType) if compatible
	}
	// Verify service type compatibility if needed, for now using PPPoE flow as base.
	if input.ServiceType == model.ServiceTypePPPoE {
		args["service"] = "pppoe"
	} else {
		// Fallback or specific logic for other types.
		// For now, to ensure 'service' arg is valid for PPP secret:
		args["service"] = "any" // or handle separately
	}

	// Optional: Add comment with customer name
	if input.Name != "" {
		args["comment"] = input.Name
	}

	// Optional: Add local/remote address if configured in profile
	if profile.PPPoE != nil {
		if profile.PPPoE.LocalAddress != nil {
			args["local-address"] = *profile.PPPoE.LocalAddress
		}
		if profile.PPPoE.RemoteAddress != nil {
			args["remote-address"] = *profile.PPPoE.RemoteAddress
		}
	}

	// Call MikroTik API to create PPPoE secret
	// This MUST be done before DB insert to get the ID
	reply, err := client.RunArgs("/ppp/secret/add", args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ppp secret in mikrotik")
	}

	// Extract mikrotik object ID from reply
	var mikrotikObjectID string
	if reply.Done != nil && reply.Done.Map != nil {
		if ret, ok := reply.Done.Map["ret"]; ok {
			mikrotikObjectID = ret
		} else if after, ok := reply.Done.Map["after"]; ok {
			mikrotikObjectID = after
		}
	}
	if mikrotikObjectID == "" {
		// Immediate cleanup if ID missing
		// Try to delete by name just in case it was created? Hard without ID.
		// Usually 'ret' is present on creation.
		return nil, fmt.Errorf("failed to get mikrotik object id from response")
	}

	// Begin database transaction
	result, err := d.databasePort.DoInTransaction(ctx, func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Insert to customers table with MikroTik ID
		customer, err := txDB.Customer().CreateCustomer(ctx, input, activeMikrotik.ID, mikrotikObjectID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create customer")
		}

		// 2. Insert to customer_services table
		_, err = txDB.Customer().CreateCustomerService(
			ctx,
			customer.ID,
			input.ProfileID,
			profile.Price,
			profile.TaxRate,
			*input.StartDate,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create customer service")
		}

		// 3. Get complete customer with service
		customerWithService, err := txDB.Customer().GetByID(ctx, customer.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get created customer")
		}

		return customerWithService, nil
	})

	if err != nil {
		// Transaction failed, rollback MikroTik creation
		// We use the client created outside transaction
		_, _ = client.RunArgs("/ppp/secret/remove", map[string]string{
			".id": mikrotikObjectID,
		})
		return nil, err
	}

	return result.(*model.CustomerWithService), nil
}

func (d *customerDomain) GetCustomer(ctx context.Context, id string) (*model.CustomerWithService, error) {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	customer, err := d.databasePort.Customer().GetByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	return customer, nil
}

func (d *customerDomain) ListCustomers(ctx context.Context) ([]model.CustomerWithService, error) { // Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	customers, err := d.databasePort.Customer().List(ctx, activeMikrotik.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list customers")
	}

	return customers, nil
}

func (d *customerDomain) UpdateCustomer(ctx context.Context, id string, input model.CustomerInput) (*model.CustomerWithService, error) {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	model.PrepareCustomerInput(&input)

	// Get existing customer to get mikrotik_object_id
	existing, err := d.databasePort.Customer().GetByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get existing customer")
	}

	if existing.Service == nil {
		return nil, fmt.Errorf("customer has no active service")
	}
	if existing.MikrotikObjectID == "" {
		return nil, fmt.Errorf("customer has no mikrotik object id")
	}

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	// Validate profile exists if changed or needed
	profile, err := d.databasePort.Profile().GetByMikrotikID(ctx, activeMikrotik.ID, input.ProfileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}
	if profile == nil {
		return nil, fmt.Errorf("profile not found")
	}

	result, err := d.databasePort.DoInTransaction(ctx, func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Update database
		err := txDB.Customer().Update(ctx, customerID, input, profile.Price, profile.TaxRate)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update customer in database")
		}

		// 2. Update MikroTik
		client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
		}
		defer client.Close()

		// Prepare update parameters
		args := map[string]string{
			".id":      existing.MikrotikObjectID,
			"name":     input.Username,
			"password": input.Password,
			"profile":  profile.Name,
		}

		if input.Name != "" {
			args["comment"] = input.Name
		}

		_, err = client.RunArgs("/ppp/secret/set", args)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update ppp secret in mikrotik")
		}

		// 3. Get updated customer
		updated, err := txDB.Customer().GetByID(ctx, customerID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get updated customer")
		}

		return updated, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.CustomerWithService), nil
}

func (d *customerDomain) DeleteCustomer(ctx context.Context, id string) error {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid customer id")
	}

	// Get existing customer to get mikrotik_object_id
	existing, err := d.databasePort.Customer().GetByID(ctx, customerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get existing customer")
	}

	if existing.MikrotikObjectID != "" {
		// Valid to delete from MikroTik
	} else {
		// Just delete from DB if no MikroTik ID (maybe failed creation previously)
	}

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	_, err = d.databasePort.DoInTransaction(ctx, func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Delete from MikroTik if ID exists
		if existing.MikrotikObjectID != "" {
			client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
			}
			defer client.Close()

			_, err = client.RunArgs("/ppp/secret/remove", map[string]string{
				".id": existing.MikrotikObjectID,
			})
			if err != nil {
				// We log error but maybe proceed? Or fail?
				// Better fail to ensure sync, or if not found (already deleted) then ok.
				// For now propagate error.
				return nil, stacktrace.Propagate(err, "failed to delete ppp secret from mikrotik")
			}
		}

		// 2. Delete from database
		err = txDB.Customer().Delete(ctx, customerID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to delete customer from database")
		}

		return nil, nil
	})

	return err
}

func (d *customerDomain) HandlePPPoEUp(ctx context.Context, input model.PPPoEUpInput) error { // Find customer by username
	customer, err := d.databasePort.Customer().GetByPPPoEUsername(ctx, input.User)
	if err != nil {
		// Log warning but don't error out completely if user not found?
		// Or return error and let handler decide code.
		return stacktrace.Propagate(err, "failed to find customer for pppoe up callback")
	}

	// Update status
	status := model.CustomerStatusActive
	err = d.databasePort.Customer().UpdateStatus(ctx, customer.ID, status, &input.IPAddress, &input.MacAddress, &input.Interface)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update customer status")
	}

	// Publish event
	eventData := fmt.Sprintf(`{"type":"pppoe_event","status":"connected","customer_id":"%s","name":"%s","ip":"%s","interface":"%s"}`,
		customer.ID, customer.Name, input.IPAddress, input.Interface)

	err = d.cachePort.PubSub().Publish("mikrotik:events", eventData)
	if err != nil {
		// Log error but treat as success for the callback
		// In a real logger we would log.Warn
		// Here we just return error wrapped? Or return nil?
		// User code logged warning and returned success.
		// Propagate or swallow?
		// "Failed to publish Redis event: %v"
		// I will propagate it for now, handler can ignore.
		return stacktrace.Propagate(err, "failed to publish redis event")
	}

	return nil
}

func (d *customerDomain) HandlePPPoEDown(ctx context.Context, input model.PPPoEDownInput) error { // Find customer by username
	customer, err := d.databasePort.Customer().GetByPPPoEUsername(ctx, input.User)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer for pppoe down callback")
	}

	// Update status to inactive
	status := model.CustomerStatusInactive
	err = d.databasePort.Customer().UpdateStatus(ctx, customer.ID, status, nil, nil, nil)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update customer status")
	}

	// Publish event
	eventData := fmt.Sprintf(`{"type":"pppoe_event","status":"disconnected","customer_id":"%s","name":"%s"}`,
		customer.ID, customer.Name)

	err = d.cachePort.PubSub().Publish("mikrotik:events", eventData)
	if err != nil {
		return stacktrace.Propagate(err, "failed to publish redis event")
	}

	return nil
}

