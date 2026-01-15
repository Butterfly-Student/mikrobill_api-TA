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

const (
	tableCustomers        = "customers"
	tableCustomerServices = "customer_services"
)

type customerAdapter struct {
	db outbound_port.DatabaseExecutor
}

func NewCustomerAdapter(
	db outbound_port.DatabaseExecutor,
) outbound_port.CustomerDatabasePort {
	return &customerAdapter{
		db: db,
	}
}

func (a *customerAdapter) CreateCustomer(ctx context.Context, input model.CustomerInput, mikrotikID uuid.UUID, mikrotikObjectID string) (*model.Customer, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	record := goqu.Record{
		"tenant_id":          tenantID,
		"mikrotik_id":        mikrotikID,
		"username":           input.Username,
		"name":               input.Name,
		"phone":              input.Phone,
		"email":              input.Email,
		"address":            input.Address,
		"mikrotik_object_id": mikrotikObjectID,
		"service_type":       input.ServiceType,
		"status":             model.CustomerStatusInactive,
		"auto_suspension":    input.AutoSuspension,
		"billing_day":        input.BillingDay,
		"customer_notes":     input.CustomerNotes,
		"join_date":          time.Now(),
	}

	query, _, err := goqu.Dialect("postgres").
		Insert(tableCustomers).
		Rows(record).
		Returning("*").
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build insert customer query")
	}

	var result model.Customer
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.MikrotikID,
		&result.PackageID,
		&result.Username,
		&result.Name,
		&result.Phone,
		&result.Email,
		&result.Address,
		&result.MikrotikObjectID,
		&result.ServiceType,
		&result.AssignedIP, // Nullable
		&result.MacAddress, // Nullable
		&result.Interface,  // Nullable
		&result.LastOnline, // Nullable
		&result.LastIP,     // Nullable
		&result.Status,
		&result.AutoSuspension,
		&result.BillingDay,
		&result.JoinDate,
		&result.CustomerNotes,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to insert customer")
	}

	return &result, nil
}

func (a *customerAdapter) CreateCustomerService(ctx context.Context, customerID, profileID uuid.UUID, price, taxRate float64, startDate time.Time) (*model.CustomerService, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	record := goqu.Record{
		"tenant_id":   tenantID,
		"customer_id": customerID,
		"profile_id":  profileID,
		"price":       price,
		"tax_rate":    taxRate,
		"start_date":  startDate,
		"status":      model.ServiceStatusActive,
	}

	query, _, err := goqu.Dialect("postgres").
		Insert(tableCustomerServices).
		Rows(record).
		Returning("*").
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build insert service query")
	}

	var result model.CustomerService
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.CustomerID,
		&result.ProfileID,
		&result.Price,
		&result.TaxRate,
		&result.StartDate,
		&result.EndDate,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to insert customer service")
	}

	return &result, nil
}

func (a *customerAdapter) UpdateMikrotikObjectID(ctx context.Context, customerID uuid.UUID, objectID string) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	query, _, err := goqu.Dialect("postgres").
		Update(tableCustomers).
		Set(goqu.Record{"mikrotik_object_id": objectID}).
		Where(goqu.Ex{"id": customerID, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build update query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update mikrotik object id")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("customer not found")
	}

	return nil
}

func (a *customerAdapter) GetByID(ctx context.Context, id uuid.UUID) (*model.CustomerWithService, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Query customer
	customerQuery, _, err := goqu.Dialect("postgres").
		From(tableCustomers).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build customer query")
	}

	var customer model.Customer
	err = a.db.QueryRowContext(ctx, customerQuery).Scan(
		&customer.ID,
		&customer.MikrotikID,
		&customer.PackageID,
		&customer.Username,
		&customer.Name,
		&customer.Phone,
		&customer.Email,
		&customer.Address,
		&customer.MikrotikObjectID,
		&customer.ServiceType,
		&customer.AssignedIP,
		&customer.MacAddress,
		&customer.Interface,
		&customer.LastOnline,
		&customer.LastIP,
		&customer.Status,
		&customer.AutoSuspension,
		&customer.BillingDay,
		&customer.JoinDate,
		&customer.CustomerNotes,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, stacktrace.NewError("customer not found")
		}
		return nil, stacktrace.Propagate(err, "failed to get customer")
	}

	// Query service
	serviceQuery, _, err := goqu.Dialect("postgres").
		From(tableCustomerServices).
		Where(goqu.Ex{"customer_id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build service query")
	}

	var service model.CustomerService
	err = a.db.QueryRowContext(ctx, serviceQuery).Scan(
		&service.ID,
		&service.CustomerID,
		&service.ProfileID,
		&service.Price,
		&service.TaxRate,
		&service.StartDate,
		&service.EndDate,
		&service.Status,
		&service.CreatedAt,
		&service.UpdatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, stacktrace.Propagate(err, "failed to get service")
	}

	result := &model.CustomerWithService{
		Customer: customer,
	}

	if err != sql.ErrNoRows {
		result.Service = &service
	}

	return result, nil
}

func (a *customerAdapter) GetByUsername(ctx context.Context, mikrotikID uuid.UUID, username string) (*model.Customer, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	query, _, err := goqu.Dialect("postgres").
		From(tableCustomers).
		Where(goqu.Ex{"mikrotik_id": mikrotikID, "username": username, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build query")
	}

	var result model.Customer
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.MikrotikID,
		&result.PackageID,
		&result.Username,
		&result.Name,
		&result.Phone,
		&result.Email,
		&result.Address,
		&result.MikrotikObjectID,
		&result.ServiceType,
		&result.AssignedIP,
		&result.MacAddress,
		&result.Interface,
		&result.LastOnline,
		&result.LastIP,
		&result.Status,
		&result.AutoSuspension,
		&result.BillingDay,
		&result.JoinDate,
		&result.CustomerNotes,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found is valid
		}
		return nil, stacktrace.Propagate(err, "failed to get customer by username")
	}

	return &result, nil
}

func (a *customerAdapter) List(ctx context.Context, mikrotikID uuid.UUID) ([]model.CustomerWithService, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Query all customers for this MikroTik
	customersQuery, _, err := goqu.Dialect("postgres").
		From(tableCustomers).
		Where(goqu.Ex{"mikrotik_id": mikrotikID, "tenant_id": tenantID}).
		Order(goqu.I("created_at").Desc()).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build customers query")
	}

	rows, err := a.db.QueryContext(ctx, customersQuery)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to query customers")
	}
	defer rows.Close()

	var customers []model.Customer
	for rows.Next() {
		var customer model.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.MikrotikID,
			&customer.PackageID,
			&customer.Username,
			&customer.Name,
			&customer.Phone,
			&customer.Email,
			&customer.Address,
			&customer.MikrotikObjectID,
			&customer.ServiceType,
			&customer.AssignedIP,
			&customer.MacAddress,
			&customer.Interface,
			&customer.LastOnline,
			&customer.LastIP,
			&customer.Status,
			&customer.AutoSuspension,
			&customer.BillingDay,
			&customer.JoinDate,
			&customer.CustomerNotes,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to scan customer")
		}
		customers = append(customers, customer)
	}

	// For each customer, get service
	var result []model.CustomerWithService
	for _, customer := range customers {
		serviceQuery, _, err := goqu.Dialect("postgres").
			From(tableCustomerServices).
			Where(goqu.Ex{"customer_id": customer.ID, "tenant_id": tenantID}).
			ToSQL()
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to build service query")
		}

		var service model.CustomerService
		err = a.db.QueryRowContext(ctx, serviceQuery).Scan(
			&service.ID,
			&service.CustomerID,
			&service.ProfileID,
			&service.Price,
			&service.TaxRate,
			&service.StartDate,
			&service.EndDate,
			&service.Status,
			&service.CreatedAt,
			&service.UpdatedAt,
		)

		customerWithService := model.CustomerWithService{
			Customer: customer,
		}

		if err != sql.ErrNoRows {
			customerWithService.Service = &service
		}

		result = append(result, customerWithService)
	}

	return result, nil
}

func (a *customerAdapter) Update(ctx context.Context, id uuid.UUID, input model.CustomerInput, price, taxRate float64) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Update customer
	customerUpdate := goqu.Record{
		"username":        input.Username,
		"name":            input.Name,
		"phone":           input.Phone,
		"email":           input.Email,
		"address":         input.Address,
		"service_type":    input.ServiceType,
		"auto_suspension": input.AutoSuspension,
		"billing_day":     input.BillingDay,
		"customer_notes":  input.CustomerNotes,
	}

	customerQuery, _, err := goqu.Dialect("postgres").
		Update(tableCustomers).
		Set(customerUpdate).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build customer update query")
	}

	result, err := a.db.ExecContext(ctx, customerQuery)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update customer")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("customer not found")
	}

	// Update service fields
	serviceUpdate := goqu.Record{
		"profile_id": input.ProfileID,
		"price":      price,
		"tax_rate":   taxRate,
	}
	if input.StartDate != nil {
		serviceUpdate["start_date"] = *input.StartDate
	}

	serviceUpdateQuery, _, err := goqu.Dialect("postgres").
		Update(tableCustomerServices).
		Set(serviceUpdate).
		Where(goqu.Ex{"customer_id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build service update query")
	}

	_, err = a.db.ExecContext(ctx, serviceUpdateQuery)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update service")
	}

	return nil
}

func (a *customerAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	// Delete customer (service will be cascade deleted)
	query, _, err := goqu.Dialect("postgres").
		Delete(tableCustomers).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build delete query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete customer")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("customer not found")
	}

	return nil
}

func (a *customerAdapter) GetByPPPoEUsername(ctx context.Context, username string) (*model.Customer, error) {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	query, _, err := goqu.Dialect("postgres").
		From(tableCustomers).
		Where(goqu.Ex{"username": username, "tenant_id": tenantID}).
		Limit(1).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build query")
	}

	var result model.Customer
	err = a.db.QueryRowContext(ctx, query).Scan(
		&result.ID,
		&result.MikrotikID,
		&result.PackageID,
		&result.Username,
		&result.Name,
		&result.Phone,
		&result.Email,
		&result.Address,
		&result.MikrotikObjectID,
		&result.ServiceType,
		&result.AssignedIP,
		&result.MacAddress,
		&result.Interface,
		&result.LastOnline,
		&result.LastIP,
		&result.Status,
		&result.AutoSuspension,
		&result.BillingDay,
		&result.JoinDate,
		&result.CustomerNotes,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, stacktrace.NewError("customer not found")
		}
		return nil, stacktrace.Propagate(err, "failed to get customer by pppoe username")
	}

	return &result, nil
}

func (a *customerAdapter) UpdateStatus(ctx context.Context, id uuid.UUID, status model.CustomerStatus, ip, mac, interfaceName *string) error {
	tenantID, err := contextutil.GetTenantID(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get tenant ID from context")
	}

	record := goqu.Record{
		"status": status,
	}
	if ip != nil {
		record["assigned_ip"] = *ip
		record["last_ip"] = *ip
		record["last_online"] = time.Now()
	}
	if mac != nil {
		record["mac_address"] = *mac
	}
	if interfaceName != nil {
		record["interface"] = *interfaceName
	}

	query, _, err := goqu.Dialect("postgres").
		Update(tableCustomers).
		Set(record).
		Where(goqu.Ex{"id": id, "tenant_id": tenantID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build update status query")
	}

	result, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update customer status")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("customer not found")
	}

	return nil
}
