package ppp

import (
	"context"
	"fmt"

	"github.com/palantir/stacktrace"

	"MikrOps/internal/model"
)

// --- Active PPP Connections ---

func (d *PPPDomain) MikrotikListActive(ctx context.Context) ([]model.PPPActive, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.Run("/ppp/active/print")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list active ppp connections")
	}

	var active []model.PPPActive
	for _, re := range reply.Re {
		m := re.Map
		active = append(active, model.PPPActive{
			ID:       m[".id"],
			Name:     m["name"],
			Address:  m["address"],
			Uptime:   m["uptime"],
			Encoding: m["encoding"],
			CallerID: m["caller-id"],
			Service:  m["service"],
		})
	}
	return active, nil
}

func (d *PPPDomain) MikrotikGetActive(ctx context.Context, id string) (*model.PPPActive, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.RunArgs("/ppp/active/print", map[string]string{
		"?.id": id,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active ppp connection")
	}
	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("active ppp connection not found")
	}

	m := reply.Re[0].Map
	return &model.PPPActive{
		ID:       m[".id"],
		Name:     m["name"],
		Address:  m["address"],
		Uptime:   m["uptime"],
		Encoding: m["encoding"],
		CallerID: m["caller-id"],
		Service:  m["service"],
	}, nil
}
