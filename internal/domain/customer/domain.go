package customer

import (
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
}

func NewCustomerDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
) inbound_port.CustomerDomain {
	return &customerDomain{
		databasePort:          databasePort,
		mikrotikClientFactory: mikrotikClientFactory,
	}
}

func (d *customerDomain) CreateCustomer(ctx any, input model.CustomerInput) (*model.CustomerWithService, error) {
	// Validate and prepare input
	model.PrepareCustomerInput(&input)

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	// Validate profile exists
	profile, err := d.databasePort.Profile().GetByMikrotikID(activeMikrotik.ID, input.ProfileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}
	if profile == nil {
		return nil, fmt.Errorf("profile not found")
	}

	// Check if customer already exists
	existingCustomer, err := d.databasePort.Customer().GetByUsername(activeMikrotik.ID, input.Username)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing customer")
	}
	if existingCustomer != nil {
		return nil, fmt.Errorf("customer with username '%s' already exists", input.Username)
	}

	// Begin database transaction
	result, err := d.databasePort.DoInTransaction(func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Insert to customers table
		customer, err := txDB.Customer().CreateCustomer(input, activeMikrotik.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create customer")
		}

		// 2. Insert to customer_services table
		service, err := txDB.Customer().CreateCustomerService(
			customer.ID,
			input.ProfileID,
			input.Price,
			*input.TaxRate,
			*input.StartDate,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create customer service")
		}

		// 3. Create MikroTik client
		client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
		}
		defer client.Close()

		// 4. Prepare MikroTik PPP Secret parameters
		args := map[string]string{
			"name":     input.Username,
			"password": input.Password,
			"profile":  profile.Name,
			"service":  "pppoe",
		}

		// Optional: Add comment with customer full name
		if input.FullName != "" {
			args["comment"] = input.FullName
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

		// 5. Call MikroTik API to create PPPoE secret
		reply, err := client.RunArgs("/ppp/secret/add", args)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create ppp secret in mikrotik")
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

		// 7. Update mikrotik_object_id in customer_services
		err = txDB.Customer().UpdateServiceMikrotikObjectID(service.ID, mikrotikObjectID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update mikrotik object id")
		}

		// 8. Get complete customer with service
		customerWithService, err := txDB.Customer().GetByID(customer.ID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get created customer")
		}

		return customerWithService, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.CustomerWithService), nil
}

func (d *customerDomain) GetCustomer(ctx any, id string) (*model.CustomerWithService, error) {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	customer, err := d.databasePort.Customer().GetByID(customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	return customer, nil
}

func (d *customerDomain) ListCustomers(ctx any) ([]model.CustomerWithService, error) {
	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	customers, err := d.databasePort.Customer().List(activeMikrotik.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list customers")
	}

	return customers, nil
}

func (d *customerDomain) UpdateCustomer(ctx any, id string, input model.CustomerInput) (*model.CustomerWithService, error) {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid customer id")
	}

	model.PrepareCustomerInput(&input)

	// Get existing customer to get mikrotik_object_id
	existing, err := d.databasePort.Customer().GetByID(customerID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get existing customer")
	}

	if existing.Service == nil {
		return nil, fmt.Errorf("customer has no active service")
	}
	if existing.Service.MikrotikObjectID == nil || *existing.Service.MikrotikObjectID == "" {
		return nil, fmt.Errorf("customer has no mikrotik object id")
	}

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	// Validate profile exists if changed or needed
	profile, err := d.databasePort.Profile().GetByMikrotikID(activeMikrotik.ID, input.ProfileID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get profile")
	}
	if profile == nil {
		return nil, fmt.Errorf("profile not found")
	}

	result, err := d.databasePort.DoInTransaction(func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Update database
		err := txDB.Customer().Update(customerID, input)
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
			".id":      *existing.Service.MikrotikObjectID,
			"name":     input.Username,
			"password": input.Password,
			"profile":  profile.Name,
		}

		if input.FullName != "" {
			args["comment"] = input.FullName
		}

		_, err = client.RunArgs("/ppp/secret/set", args)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to update ppp secret in mikrotik")
		}

		// 3. Get updated customer
		updated, err := txDB.Customer().GetByID(customerID)
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

func (d *customerDomain) DeleteCustomer(ctx any, id string) error {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return stacktrace.Propagate(err, "invalid customer id")
	}

	// Get existing customer to get mikrotik_object_id
	existing, err := d.databasePort.Customer().GetByID(customerID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get existing customer")
	}

	if existing.Service != nil && existing.Service.MikrotikObjectID != nil && *existing.Service.MikrotikObjectID != "" {
		// Valid to delete from MikroTik
	} else {
		// Just delete from DB if no MikroTik ID (maybe failed creation previously)
	}

	// Get active mikrotik
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	_, err = d.databasePort.DoInTransaction(func(txDB outbound_port.DatabasePort) (interface{}, error) {
		// 1. Delete from MikroTik if ID exists
		if existing.Service != nil && existing.Service.MikrotikObjectID != nil && *existing.Service.MikrotikObjectID != "" {
			client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
			if err != nil {
				return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
			}
			defer client.Close()

			_, err = client.RunArgs("/ppp/secret/remove", map[string]string{
				".id": *existing.Service.MikrotikObjectID,
			})
			if err != nil {
				// We log error but maybe proceed? Or fail?
				// Better fail to ensure sync, or if not found (already deleted) then ok.
				// For now propagate error.
				return nil, stacktrace.Propagate(err, "failed to delete ppp secret from mikrotik")
			}
		}

		// 2. Delete from database
		err = txDB.Customer().Delete(customerID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to delete customer from database")
		}

		return nil, nil
	})

	return err
}
