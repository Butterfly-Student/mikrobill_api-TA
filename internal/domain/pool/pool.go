package pool

import (
	"context"
	"fmt"

	"github.com/palantir/stacktrace"

	"MikrOps/internal/model"
)

// --- IP Pool CRUD ---

func (d *PoolDomain) MikrotikCreatePool(ctx context.Context, input model.IPPoolInput) (*model.IPPool, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	args := map[string]string{
		"name":   input.Name,
		"ranges": input.Ranges,
	}
	if input.NextPool != "" {
		args["next-pool"] = input.NextPool
	}

	_, err = client.RunArgs("/ip/pool/add", args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ip pool")
	}

	return &model.IPPool{
		Name:     input.Name,
		Ranges:   input.Ranges,
		NextPool: input.NextPool,
	}, nil
}

func (d *PoolDomain) MikrotikGetPool(ctx context.Context, id string) (*model.IPPool, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.RunArgs("/ip/pool/print", map[string]string{
		"?.id": id,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get ip pool")
	}
	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("ip pool not found")
	}

	m := reply.Re[0].Map
	return &model.IPPool{
		ID:       m[".id"],
		Name:     m["name"],
		Ranges:   m["ranges"],
		NextPool: m["next-pool"],
	}, nil
}

func (d *PoolDomain) MikrotikUpdatePool(ctx context.Context, id string, input model.IPPoolUpdateInput) (*model.IPPool, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	args := map[string]string{".id": id}
	if input.Name != nil {
		args["name"] = *input.Name
	}
	if input.Ranges != nil {
		args["ranges"] = *input.Ranges
	}
	if input.NextPool != nil {
		args["next-pool"] = *input.NextPool
	}

	_, err = client.RunArgs("/ip/pool/set", args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to update ip pool")
	}

	return nil, nil
}

func (d *PoolDomain) MikrotikDeletePool(ctx context.Context, id string) error {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.RunArgs("/ip/pool/remove", map[string]string{
		".id": id,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete ip pool")
	}
	return nil
}

func (d *PoolDomain) MikrotikListPools(ctx context.Context) ([]model.IPPool, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.Run("/ip/pool/print")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list ip pools")
	}

	var pools []model.IPPool
	for _, re := range reply.Re {
		m := re.Map
		pools = append(pools, model.IPPool{
			ID:       m[".id"],
			Name:     m["name"],
			Ranges:   m["ranges"],
			NextPool: m["next-pool"],
		})
	}
	return pools, nil
}
