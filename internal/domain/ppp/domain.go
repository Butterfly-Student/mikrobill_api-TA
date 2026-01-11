package ppp

import (
	"fmt"

	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	inbound_port "prabogo/internal/port/inbound"
	outbound_port "prabogo/internal/port/outbound"
)

type pppDomain struct {
	databasePort          outbound_port.DatabasePort
	mikrotikClientFactory outbound_port.MikrotikClientFactory
}

func NewPPPDomain(
	databasePort outbound_port.DatabasePort,
	mikrotikClientFactory outbound_port.MikrotikClientFactory,
) inbound_port.PPPDomain {
	return &pppDomain{
		databasePort:          databasePort,
		mikrotikClientFactory: mikrotikClientFactory,
	}
}

func (d *pppDomain) getActiveClient() (outbound_port.MikrotikClientPort, error) {
	// Get active mikrotik from DB
	activeMikrotik, err := d.databasePort.Mikrotik().GetActiveMikrotik()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}
	if activeMikrotik == nil {
		return nil, fmt.Errorf("no active mikrotik found")
	}

	// Create client
	client, err := d.mikrotikClientFactory.NewClient(activeMikrotik)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mikrotik client")
	}
	return client, nil
}

// --- Secrets ---

func (d *pppDomain) CreateSecret(ctx any, input model.PPPSecretInput) (*model.PPPSecret, error) {
	client, err := d.getActiveClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	_, err = client.RunArgs("/ppp/secret/add", map[string]string{
		"name":           input.Name,
		"password":       input.Password,
		"profile":        input.Profile,
		"service":        input.Service,
		"local-address":  input.LocalAddress,
		"remote-address": input.RemoteAddress,
		"comment":        input.Comment,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ppp secret")
	}

	// Return mock object as we can't easily get the ID back immediately from RouterOS add
	// Or we could query it back. For now returning input as success.
	return &model.PPPSecret{
		Name:    input.Name,
		Profile: input.Profile,
	}, nil
}

func (d *pppDomain) GetSecret(ctx any, id string) (*model.PPPSecret, error) {
	client, err := d.getActiveClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.RunArgs("/ppp/secret/print", map[string]string{
		"?.id": id,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get ppp secret")
	}
	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("ppp secret not found")
	}

	r := reply.Re[0].Map
	return &model.PPPSecret{
		Name:     r["name"],
		Profile:  r["profile"],
		Service:  r["service"],
		Disabled: r["disabled"] == "true",
	}, nil
}

func (d *pppDomain) UpdateSecret(ctx any, id string, input model.PPPSecretUpdateInput) (*model.PPPSecret, error) {
	client, err := d.getActiveClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	args := map[string]string{".id": id}
	if input.Password != nil {
		args["password"] = *input.Password
	}
	if input.Profile != nil {
		args["profile"] = *input.Profile
	}
	if input.Service != nil {
		args["service"] = *input.Service
	}
	// ... other fields

	_, err = client.RunArgs("/ppp/secret/set", args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to update ppp secret")
	}

	return nil, nil // Return nil or updated object
}

func (d *pppDomain) DeleteSecret(ctx any, id string) error {
	client, err := d.getActiveClient()
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.RunArgs("/ppp/secret/remove", map[string]string{
		".id": id,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete ppp secret")
	}
	return nil
}

func (d *pppDomain) ListSecrets(ctx any) ([]model.PPPSecret, error) {
	client, err := d.getActiveClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.Run("/ppp/secret/print")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list ppp secrets")
	}

	var secrets []model.PPPSecret
	for _, re := range reply.Re {
		m := re.Map
		secrets = append(secrets, model.PPPSecret{
			Name:     m["name"],
			Profile:  m["profile"],
			Service:  m["service"],
			Disabled: m["disabled"] == "true",
		})
	}
	return secrets, nil
}

// --- Profiles ---

func (d *pppDomain) CreateProfile(ctx any, input model.PPPProfileInput) (*model.PPPProfile, error) {
	client, err := d.getActiveClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	_, err = client.RunArgs("/ppp/profile/add", map[string]string{
		"name":           input.Name,
		"local-address":  input.LocalAddress,
		"remote-address": input.RemoteAddress,
		"rate-limit":     input.RateLimit,
		"only-one":       input.OnlyOne,
		"comment":        input.Comment,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ppp profile")
	}

	return &model.PPPProfile{Name: input.Name}, nil
}

func (d *pppDomain) GetProfile(ctx any, id string) (*model.PPPProfile, error) {
	client, err := d.getActiveClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.RunArgs("/ppp/profile/print", map[string]string{
		"?.id": id,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get ppp profile")
	}
	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("ppp profile not found")
	}

	r := reply.Re[0].Map
	return &model.PPPProfile{
		Name:      r["name"],
		RateLimit: r["rate-limit"],
	}, nil
}

func (d *pppDomain) UpdateProfile(ctx any, id string, input model.PPPProfileInput) (*model.PPPProfile, error) {
	client, err := d.getActiveClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	args := map[string]string{".id": id}
	args["name"] = input.Name
	// ... other fields

	_, err = client.RunArgs("/ppp/profile/set", args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to update ppp profile")
	}

	return nil, nil
}

func (d *pppDomain) DeleteProfile(ctx any, id string) error {
	client, err := d.getActiveClient()
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.RunArgs("/ppp/profile/remove", map[string]string{
		".id": id,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete ppp profile")
	}
	return nil
}

func (d *pppDomain) ListProfiles(ctx any) ([]model.PPPProfile, error) {
	client, err := d.getActiveClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.Run("/ppp/profile/print")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list ppp profiles")
	}

	var profiles []model.PPPProfile
	for _, re := range reply.Re {
		m := re.Map
		profiles = append(profiles, model.PPPProfile{
			Name:      m["name"],
			RateLimit: m["rate-limit"],
		})
	}
	return profiles, nil
}
