package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"MikrOps/internal/model"
	inbound_port "MikrOps/internal/port/inbound"
	outbound_port "MikrOps/internal/port/outbound"
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

func (d *customerDomain) CreateCustomer(ctx context.Context, input model.CreateCustomerRequest) (*model.Customer, error) {
	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	// Parse IDs
	mikrotikID, err := uuid.Parse(activeMikrotik.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid mikrotik id")
	}

	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid profile id")
	}

	// Validate profile exists
	profile, err := d.databasePort.Profile().GetByMikrotikID(ctx, mikrotikID, profileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}
	if profile == nil {
		return nil, fmt.Errorf("profile not found")
	}

	// Check if customer already exists
	existingCustomer, err := d.databasePort.Customer().GetByUsername(ctx, mikrotikID, input.Username)
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
	args := map[string]string{
		"name":     input.Username,
		"password": input.Password,
		"profile":  profile.Name,
		"service":  "pppoe",
	}

	if input.ServiceType == model.ServiceTypePPPoE {
		args["service"] = "pppoe"
	} else {
		args["service"] = "any"
	}

	if input.Name != "" {
		args["comment"] = input.Name
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

	// Call MikroTik API to create PPPoE secret
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
		return nil, fmt.Errorf("failed to get mikrotik object id from response")
	}

	// Begin database transaction
	result, err := d.databasePort.DoInTransaction(ctx, func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Insert to customers table with MikroTik ID
		customer, err := txDB.Customer().CreateCustomer(ctx, input, mikrotikID, mikrotikObjectID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create customer")
		}

		// 2. Insert to customer_services table
		startDate := time.Now()
		if input.StartDate != nil {
			startDate = *input.StartDate
		}

		custID, err := uuid.Parse(customer.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "invalid customer id")
		}

		_, err = txDB.Customer().CreateCustomerService(
			ctx,
			custID,
			profileID,
			profile.Price,
			profile.TaxRate,
			startDate,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create customer service")
		}

		// 3. Get complete customer
		customerWithService, err := txDB.Customer().GetByID(ctx, custID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get created customer")
		}

		return customerWithService, nil
	})

	if err != nil {
		// Transaction failed, rollback MikroTik creation
		_, _ = client.RunArgs("/ppp/secret/remove", map[string]string{
			".id": mikrotikObjectID,
		})
		return nil, err
	}

	return result.(*model.Customer), nil
}

func (d *customerDomain) GetCustomer(ctx context.Context, id string) (*model.Customer, error) {
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

func (d *customerDomain) ListCustomers(ctx context.Context) ([]model.Customer, error) {
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

	customers, err := d.databasePort.Customer().List(ctx, mikrotikID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list customers")
	}

	return customers, nil
}

func (d *customerDomain) UpdateCustomer(ctx context.Context, id string, input model.CreateCustomerRequest) (*model.Customer, error) {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	// Get existing customer to get mikrot ik_object_id
	existing, err := d.databasePort.Customer().GetByID(ctx, customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get existing customer")
	}

	if existing.MikrotikObjectID == nil || *existing.MikrotikObjectID == "" {
		return nil, fmt.Errorf("customer has no mikrotik object id")
	}

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	mikrotikID, err := uuid.Parse(activeMikrotik.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid mikrotik id")
	}

	profileID, err := uuid.Parse(input.ProfileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid profile id")
	}

	// Validate profile exists
	profile, err := d.databasePort.Profile().GetByMikrotikID(ctx, mikrotikID, profileID)
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
			".id":      *existing.MikrotikObjectID,
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

	return result.(*model.Customer), nil
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

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	_, err = d.databasePort.DoInTransaction(ctx, func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Delete from MikroTik if ID exists
		if existing.MikrotikObjectID != nil && *existing.MikrotikObjectID != "" {
			client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
			}
			defer client.Close()

			_, err = client.RunArgs("/ppp/secret/remove", map[string]string{
				".id": *existing.MikrotikObjectID,
			})
			if err != nil {
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

func (d *customerDomain) HandlePPPoEUp(ctx context.Context, input model.PPPoEEventInput) error {
	// Find customer by username
	customer, err := d.databasePort.Customer().GetByPPPoEUsername(ctx, input.Name)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer for pppoe up callback")
	}

	// Update status
	customerID, err := uuid.Parse(customer.ID)
	if err != nil {
		return stacktrace.Propagate(err, "invalid customer id")
	}

	callerID := input.CallerID
	remoteAddress := input.RemoteAddress
	interfaceName := input.Interface

	status := model.CustomerStatusActive
	err = d.databasePort.Customer().UpdateStatus(ctx, customerID, status, &remoteAddress, &callerID, &interfaceName)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update customer status")
	}

	// Publish event
	eventData := fmt.Sprintf(`{"type":"pppoe_event","status":"connected","customer_id":"%s","name":"%s","ip":"%s"}`,
		customer.ID, customer.Name, remoteAddress)

	err = d.cachePort.PubSub().Publish("mikrotik:events", eventData)
	if err != nil {
		return stacktrace.Propagate(err, "failed to publish redis event")
	}

	return nil
}

func (d *customerDomain) HandlePPPoEDown(ctx context.Context, input model.PPPoEEventInput) error {
	// Find customer by username
	customer, err := d.databasePort.Customer().GetByPPPoEUsername(ctx, input.Name)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find customer for pppoe down callback")
	}

	// Update status to inactive
	customerID, err := uuid.Parse(customer.ID)
	if err != nil {
		return stacktrace.Propagate(err, "invalid customer id")
	}

	status := model.CustomerStatusInactive
	err = d.databasePort.Customer().UpdateStatus(ctx, customerID, status, nil, nil, nil)
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

