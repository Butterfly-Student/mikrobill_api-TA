package ppp

import (
	"context"
	"fmt"

	"github.com/palantir/stacktrace"

	"MikrOps/internal/model"
)

// --- Secrets ---

func (d *PPPDomain) MikrotikCreateSecret(ctx context.Context, input model.PPPSecretInput) (*model.PPPSecret, error) {
	client, err := d.getActiveClient(ctx)
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

func (d *PPPDomain) MikrotikGetSecret(ctx context.Context, id string) (*model.PPPSecret, error) {
	client, err := d.getActiveClient(ctx)
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

func (d *PPPDomain) MikrotikUpdateSecret(ctx context.Context, id string, input model.PPPSecretUpdateInput) (*model.PPPSecret, error) {
	client, err := d.getActiveClient(ctx)
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

func (d *PPPDomain) MikrotikDeleteSecret(ctx context.Context, id string) error {
	client, err := d.getActiveClient(ctx)
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

func (d *PPPDomain) MikrotikListSecrets(ctx context.Context) ([]model.PPPSecret, error) {
	client, err := d.getActiveClient(ctx)
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

