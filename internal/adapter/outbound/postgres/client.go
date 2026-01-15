package postgres_outbound_adapter

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	outbound_port "prabogo/internal/port/outbound"
)

const tableClients = "clients"

type clientAdapter struct {
	db outbound_port.DatabaseExecutor
}

func NewClientAdapter(db outbound_port.DatabaseExecutor) outbound_port.ClientDatabasePort {
	return &clientAdapter{db: db}
}

func (a *clientAdapter) Upsert(ctx context.Context, datas []model.ClientInput) error {
	if len(datas) == 0 {
		return nil
	}

	var records []goqu.Record
	for _, data := range datas {
		records = append(records, goqu.Record{
			"name":       data.Name,
			"bearer_key": data.BearerKey,
			"created_at": data.CreatedAt,
			"updated_at": data.UpdatedAt,
		})
	}

	query, _, err := goqu.Dialect("postgres").
		Insert(tableClients).
		Rows(records).
		OnConflict(goqu.DoUpdate("bearer_key", goqu.Record{
			"name":       goqu.L("EXCLUDED.name"),
			"updated_at": goqu.L("EXCLUDED.updated_at"),
		})).
		ToSQL()

	if err != nil {
		return stacktrace.Propagate(err, "failed to build upsert clients query")
	}

	_, err = a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to execute upsert clients query")
	}

	return nil
}

func (a *clientAdapter) FindByFilter(ctx context.Context, filter model.ClientFilter, lock bool) ([]model.Client, error) {
	ds := goqu.Dialect("postgres").From(tableClients)

	if len(filter.IDs) > 0 {
		ds = ds.Where(goqu.Ex{"id": filter.IDs})
	}
	if len(filter.Names) > 0 {
		ds = ds.Where(goqu.Ex{"name": filter.Names})
	}
	if len(filter.BearerKeys) > 0 {
		ds = ds.Where(goqu.Ex{"bearer_key": filter.BearerKeys})
	}

	if lock {
		ds = ds.ForUpdate(exp.Wait)
	}

	query, _, err := ds.ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build find clients query")
	}

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to execute find clients query")
	}
	defer rows.Close()

	var clients []model.Client
	for rows.Next() {
		var client model.Client
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.BearerKey,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to scan client row")
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func (a *clientAdapter) DeleteByFilter(ctx context.Context, filter model.ClientFilter) error {
	ds := goqu.Dialect("postgres").Delete(tableClients)

	if len(filter.IDs) > 0 {
		ds = ds.Where(goqu.Ex{"id": filter.IDs})
	}
	if len(filter.Names) > 0 {
		ds = ds.Where(goqu.Ex{"name": filter.Names})
	}
	if len(filter.BearerKeys) > 0 {
		ds = ds.Where(goqu.Ex{"bearer_key": filter.BearerKeys})
	}

	query, _, err := ds.ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build delete clients query")
	}

	_, err = a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to execute delete clients query")
	}

	return nil
}

func (a *clientAdapter) IsExists(ctx context.Context, bearerKey string) (bool, error) {
	query, _, err := goqu.Dialect("postgres").
		From(tableClients).
		Select(goqu.COUNT("*")).
		Where(goqu.Ex{"bearer_key": bearerKey}).
		ToSQL()

	if err != nil {
		return false, stacktrace.Propagate(err, "failed to build count client query")
	}

	var count int
	err = a.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, stacktrace.Propagate(err, "failed to execute count client query")
	}

	return count > 0, nil
}
