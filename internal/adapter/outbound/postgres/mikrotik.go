package postgres_outbound_adapter

import (
	"context"
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	outbound_port "prabogo/internal/port/outbound"
	contextutil "prabogo/utils/context"
)

const tableMikrotik = "mikrotik"

type mikrotikAdapter struct {
	db outbound_port.DatabaseExecutor
}

func NewMikrotikAdapter(
	db outbound_port.DatabaseExecutor,
) outbound_port.MikrotikDatabasePort {
	return &mikrotikAdapter{
		db: db,
	}
}

func (a *mikrotikAdapter) Create(ctx context.Context, input model.MikrotikInput) (*model.Mikrotik, error) {
	model.MikrotikPrepare(&input)

	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	record := goqu.Record{
		"tenant_id":              tenantID,
		"name":                   input.Name,
		"host":                   input.Host,
		"port":                   input.Port,
		"api_username":           input.APIUsername,
		"api_encrypted_password": input.APIEncryptedPassword,
		"keepalive":              input.Keepalive,
		"timeout":                input.Timeout,
		"location":               input.Location,
		"description":            input.Description,
		"status":                 model.MikrotikStatusOffline,
		"is_active":              false,
	}

	query, _, err := goqu.Dialect("postgres").
		Insert(tableMikrotik).
		Rows(record).
		Returning("*").
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build insert query")
	}

	var result model.Mikrotik
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.Name,
		&result.Host,
		&result.Port,
		&result.APIUsername,
		&result.APIEncryptedPassword,
		&result.Keepalive,
		&result.Timeout,
		&result.Location,
		&result.Description,
		&result.IsActive,
		&result.Status,
		&result.LastSync,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to insert mikrotik")
	}

	return &result, nil
}

func (a *mikrotikAdapter) GetByID(ctx context.Context, id uuid.UUID) (*model.Mikrotik, error) {
	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	query, _, err := goqu.Dialect("postgres").
		From(tableMikrotik).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build select query")
	}

	var result model.Mikrotik
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.Name,
		&result.Host,
		&result.Port,
		&result.APIUsername,
		&result.APIEncryptedPassword,
		&result.Keepalive,
		&result.Timeout,
		&result.Location,
		&result.Description,
		&result.IsActive,
		&result.Status,
		&result.LastSync,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, stacktrace.NewError("mikrotik not found")
		}
		return nil, stacktrace.Propagate(err, "failed to get mikrotik")
	}

	return &result, nil
}

func (a *mikrotikAdapter) List(ctx context.Context, filter model.MikrotikFilter) ([]model.Mikrotik, error) {
	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	dataset := goqu.Dialect("postgres").From(tableMikrotik).Where(goqu.Ex{"tenant_id": tenantID})
	dataset = addMikrotikFilter(dataset, filter)

	query, _, err := dataset.ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build list query")
	}

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list mikrotik")
	}
	defer rows.Close()

	var results []model.Mikrotik
	for rows.Next() {
		var result model.Mikrotik
		err := rows.Scan(
			&result.ID,
			&result.Name,
			&result.Host,
			&result.Port,
			&result.APIUsername,
			&result.APIEncryptedPassword,
			&result.Keepalive,
			&result.Timeout,
			&result.Location,
			&result.Description,
			&result.IsActive,
			&result.Status,
			&result.LastSync,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to scan mikrotik")
		}
		results = append(results, result)
	}

	return results, nil
}

func (a *mikrotikAdapter) Update(ctx context.Context, id uuid.UUID, input model.MikrotikUpdateInput) (*model.Mikrotik, error) {
	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	record := goqu.Record{}

	if input.Name != nil {
		record["name"] = *input.Name
	}
	if input.Host != nil {
		record["host"] = *input.Host
	}
	if input.Port != nil {
		record["port"] = *input.Port
	}
	if input.APIUsername != nil {
		record["api_username"] = *input.APIUsername
	}
	if input.APIEncryptedPassword != nil {
		record["api_encrypted_password"] = *input.APIEncryptedPassword
	}
	if input.Keepalive != nil {
		record["keepalive"] = *input.Keepalive
	}
	if input.Timeout != nil {
		record["timeout"] = *input.Timeout
	}
	if input.Location != nil {
		record["location"] = *input.Location
	}
	if input.Description != nil {
		record["description"] = *input.Description
	}

	if len(record) == 0 {
		return a.GetByID(ctx, id)
	}

	query, _, err := goqu.Dialect("postgres").
		Update(tableMikrotik).
		Set(record).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		Returning("*").
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build update query")
	}

	var result model.Mikrotik
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.Name,
		&result.Host,
		&result.Port,
		&result.APIUsername,
		&result.APIEncryptedPassword,
		&result.Keepalive,
		&result.Timeout,
		&result.Location,
		&result.Description,
		&result.IsActive,
		&result.Status,
		&result.LastSync,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, stacktrace.NewError("mikrotik not found")
		}
		return nil, stacktrace.Propagate(err, "failed to update mikrotik")
	}

	return &result, nil
}

func (a *mikrotikAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	query, _, err := goqu.Dialect("postgres").
		Delete(tableMikrotik).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build delete query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete mikrotik")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("mikrotik not found")
	}

	return nil
}

func (a *mikrotikAdapter) UpdateStatus(ctx context.Context, id uuid.UUID, status model.MikrotikStatus) error {
	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	query, _, err := goqu.Dialect("postgres").
		Update(tableMikrotik).
		Set(goqu.Record{"status": status}).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build update status query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update mikrotik status")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("mikrotik not found")
	}

	return nil
}

func (a *mikrotikAdapter) UpdateLastSync(ctx context.Context, id uuid.UUID) error {
	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	now := time.Now()
	query, _, err := goqu.Dialect("postgres").
		Update(tableMikrotik).
		Set(goqu.Record{"last_sync": now}).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build update last sync query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update mikrotik last sync")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("mikrotik not found")
	}

	return nil
}

func (a *mikrotikAdapter) GetActiveMikrotik(ctx context.Context) (*model.Mikrotik, error) {
	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	query, _, err := goqu.Dialect("postgres").
		From(tableMikrotik).
		Where(goqu.Ex{"is_active": true, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build get active query")
	}

	var result model.Mikrotik
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.Name,
		&result.Host,
		&result.Port,
		&result.APIUsername,
		&result.APIEncryptedPassword,
		&result.Keepalive,
		&result.Timeout,
		&result.Location,
		&result.Description,
		&result.IsActive,
		&result.Status,
		&result.LastSync,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active mikrotik is valid
		}
		return nil, stacktrace.Propagate(err, "failed to get active mikrotik")
	}

	return &result, nil
}

func (a *mikrotikAdapter) SetActive(ctx context.Context, id uuid.UUID) error {
	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Begin transaction
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return stacktrace.Propagate(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Deactivate all
	deactivateQuery, _, err := goqu.Dialect("postgres").
		Update(tableMikrotik).
		Set(goqu.Record{"is_active": false}).
		Where(goqu.Ex{"tenant_id": tenantID}). // Only deactivate for this tenant
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build deactivate query")
	}

	_, err = tx.ExecContext(ctx, deactivateQuery)
	if err != nil {
		return stacktrace.Propagate(err, "failed to deactivate all mikrotik")
	}

	// Activate the specific one
	activateQuery, _, err := goqu.Dialect("postgres").
		Update(tableMikrotik).
		Set(goqu.Record{"is_active": true}).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build activate query")
	}

	result, err := tx.ExecContext(ctx, activateQuery)
	if err != nil {
		return stacktrace.Propagate(err, "failed to activate mikrotik")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("mikrotik not found")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return stacktrace.Propagate(err, "failed to commit transaction")
	}

	return nil
}

func (a *mikrotikAdapter) DeactivateAll(ctx context.Context) error {
	// Extract tenant ID from context
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	query, _, err := goqu.Dialect("postgres").
		Update(tableMikrotik).
		Set(goqu.Record{"is_active": false}).
		Where(goqu.Ex{"tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build deactivate all query")
	}

	_, err = a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to deactivate all mikrotik")
	}

	return nil
}

func addMikrotikFilter(dataset *goqu.SelectDataset, filter model.MikrotikFilter) *goqu.SelectDataset {
	if len(filter.IDs) > 0 {
		dataset = dataset.Where(goqu.Ex{"id": filter.IDs})
	}

	if len(filter.Hosts) > 0 {
		dataset = dataset.Where(goqu.Ex{"host": filter.Hosts})
	}

	if len(filter.Statuses) > 0 {
		dataset = dataset.Where(goqu.Ex{"status": filter.Statuses})
	}

	if filter.IsActive != nil {
		dataset = dataset.Where(goqu.Ex{"is_active": *filter.IsActive})
	}

	return dataset
}
