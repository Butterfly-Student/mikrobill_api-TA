package client

import (
	"context"

	"github.com/palantir/stacktrace"

	"MikrOps/internal/model"
	outbound_port "MikrOps/internal/port/outbound"
)

type ClientDomain interface {
	Upsert(ctx context.Context, inputs []model.ClientInput) ([]model.Client, error)
	FindByFilter(ctx context.Context, filter model.ClientFilter) ([]model.Client, error)
	DeleteByFilter(ctx context.Context, filter model.ClientFilter) error
	PublishUpsert(ctx context.Context, inputs []model.ClientInput) error
	IsExists(ctx context.Context, bearerKey string) (bool, error)
}

type clientDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
}

func NewClientDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
) ClientDomain {
	return &clientDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
	}
}

func (s *clientDomain) Upsert(ctx context.Context, inputs []model.ClientInput) ([]model.Client, error) {
	if len(inputs) == 0 {
		return nil, stacktrace.NewError("inputs is empty")
	}

	var filter model.ClientFilter
	for i := range inputs {
		model.ClientPrepare(&inputs[i])
		filter.BearerKeys = append(filter.BearerKeys, inputs[i].BearerKey)
	}

	databaseClientPort := s.databasePort.Client()
	err := databaseClientPort.Upsert(ctx, inputs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "upsert client error")
	}

	results, err := databaseClientPort.FindByFilter(ctx, filter, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "find client by filter error")
	}

	return results, nil
}

func (s *clientDomain) FindByFilter(ctx context.Context, filter model.ClientFilter) ([]model.Client, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	databaseClientPort := s.databasePort.Client()
	results, err := databaseClientPort.FindByFilter(ctx, filter, false)
	if err != nil {
		return nil, stacktrace.Propagate(err, "find client by filter error")
	}

	return results, nil
}

func (s *clientDomain) DeleteByFilter(ctx context.Context, filter model.ClientFilter) error {
	if filter.IsEmpty() {
		return stacktrace.NewError("filter is empty")
	}

	databaseClientPort := s.databasePort.Client()
	err := databaseClientPort.DeleteByFilter(ctx, filter)
	if err != nil {
		return stacktrace.Propagate(err, "delete client by filter error")
	}

	return nil
}

func (s *clientDomain) PublishUpsert(ctx context.Context, inputs []model.ClientInput) error {
	if len(inputs) == 0 {
		return stacktrace.NewError("inputs is empty")
	}

	messageClientPort := s.messagePort.Client()
	err := messageClientPort.PublishUpsert(ctx, inputs)
	if err != nil {
		return stacktrace.Propagate(err, "publish upsert client error")
	}

	return nil
}

func (s *clientDomain) IsExists(ctx context.Context, bearerKey string) (bool, error) {
	if bearerKey == "" {
		return false, stacktrace.NewError("bearerKey is empty")
	}

	cacheClientPort := s.cachePort.Client()
	_, found := cacheClientPort.GetClient(ctx, bearerKey)
	if found {
		return true, nil
	}

	databaseClientPort := s.databasePort.Client()
	clientCheck, dbErr := databaseClientPort.IsExists(ctx, bearerKey)
	if dbErr != nil {
		return false, stacktrace.Propagate(dbErr, "check if client exists error")
	}

	if clientCheck {
		clients, queryErr := databaseClientPort.FindByFilter(ctx, model.ClientFilter{BearerKeys: []string{bearerKey}}, false)
		if queryErr != nil {
			return true, stacktrace.Propagate(queryErr, "find client by filter error")
		}

		if len(clients) > 0 {
			cacheErr := cacheClientPort.SetClient(ctx, bearerKey, clients[0])
			if cacheErr != nil {
				return true, stacktrace.Propagate(cacheErr, "set client to cache error")
			}
		}
	}

	return false, nil
}
