package queue

import (
	"context"
	"fmt"

	"github.com/palantir/stacktrace"

	"MikrOps/internal/model"
)

// --- Queue Simple CRUD ---

func (d *QueueDomain) MikrotikCreateQueue(ctx context.Context, input model.QueueSimpleInput) (*model.QueueSimple, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	args := map[string]string{
		"name":      input.Name,
		"target":    input.Target,
		"max-limit": input.MaxLimit,
	}
	if input.LimitAt != "" {
		args["limit-at"] = input.LimitAt
	}
	if input.Priority != "" {
		args["priority"] = input.Priority
	}

	_, err = client.RunArgs("/queue/simple/add", args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create queue simple")
	}

	return &model.QueueSimple{
		Name:     input.Name,
		Target:   input.Target,
		MaxLimit: input.MaxLimit,
		LimitAt:  input.LimitAt,
		Priority: input.Priority,
	}, nil
}

func (d *QueueDomain) MikrotikGetQueue(ctx context.Context, id string) (*model.QueueSimple, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.RunArgs("/queue/simple/print", map[string]string{
		"?.id": id,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get queue simple")
	}
	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("queue simple not found")
	}

	m := reply.Re[0].Map
	return &model.QueueSimple{
		ID:       m[".id"],
		Name:     m["name"],
		Target:   m["target"],
		MaxLimit: m["max-limit"],
		LimitAt:  m["limit-at"],
		Priority: m["priority"],
		Disabled: m["disabled"] == "true",
	}, nil
}

func (d *QueueDomain) MikrotikUpdateQueue(ctx context.Context, id string, input model.QueueSimpleUpdateInput) (*model.QueueSimple, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	args := map[string]string{".id": id}
	if input.Name != nil {
		args["name"] = *input.Name
	}
	if input.Target != nil {
		args["target"] = *input.Target
	}
	if input.MaxLimit != nil {
		args["max-limit"] = *input.MaxLimit
	}
	if input.LimitAt != nil {
		args["limit-at"] = *input.LimitAt
	}
	if input.Priority != nil {
		args["priority"] = *input.Priority
	}
	if input.Disabled != nil {
		if *input.Disabled {
			args["disabled"] = "yes"
		} else {
			args["disabled"] = "no"
		}
	}

	_, err = client.RunArgs("/queue/simple/set", args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to update queue simple")
	}

	return nil, nil
}

func (d *QueueDomain) MikrotikDeleteQueue(ctx context.Context, id string) error {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.RunArgs("/queue/simple/remove", map[string]string{
		".id": id,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete queue simple")
	}
	return nil
}

func (d *QueueDomain) MikrotikListQueues(ctx context.Context) ([]model.QueueSimple, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.Run("/queue/simple/print")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list queue simple")
	}

	var queues []model.QueueSimple
	for _, re := range reply.Re {
		m := re.Map
		queues = append(queues, model.QueueSimple{
			ID:       m[".id"],
			Name:     m["name"],
			Target:   m["target"],
			MaxLimit: m["max-limit"],
			LimitAt:  m["limit-at"],
			Priority: m["priority"],
			Disabled: m["disabled"] == "true",
		})
	}
	return queues, nil
}
