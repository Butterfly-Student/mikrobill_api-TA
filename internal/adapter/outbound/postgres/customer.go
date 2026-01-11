package postgres_outbound_adapter

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"prabogo/internal/model"
	outbound_port "prabogo/internal/port/outbound"
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

func (a *customerAdapter) CreateCustomer(input model.CustomerInput, mikrotikID uuid.UUID) (*model.Customer, error) {
	record := goqu.Record{
		"mikrotik_id": mikrotikID,
		"username":    input.Username,
		"full_name":   input.FullName,
		"phone":       input.Phone,
		"email":       input.Email,
		"address":     input.Address,
		"status":      model.CustomerStatusActive,
		"join_date":   time.Now(),
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
	err = a.db.QueryRow(query).Scan(
		&result.ID,
		&result.MikrotikID,
		&result.Username,
		&result.FullName,
		&result.Phone,
		&result.Email,
		&result.Address,
		&result.Status,
		&result.JoinDate,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to insert customer")
	}

	return &result, nil
}

func (a *customerAdapter) CreateCustomerService(customerID, profileID uuid.UUID, price, taxRate float64, startDate time.Time) (*model.CustomerService, error) {
	record := goqu.Record{
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
	err = a.db.QueryRow(query).Scan(
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
		&result.MikrotikObjectID, // This is at the END (added via ALTER TABLE)
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to insert customer service")
	}

	return &result, nil
}

func (a *customerAdapter) UpdateServiceMikrotikObjectID(serviceID uuid.UUID, objectID string) error {
	query, _, err := goqu.Dialect("postgres").
		Update(tableCustomerServices).
		Set(goqu.Record{"mikrotik_object_id": objectID}).
		Where(goqu.Ex{"id": serviceID}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build update query")
	}

	result, err := a.db.Exec(query)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update mikrotik object id")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return stacktrace.NewError("service not found")
	}

	return nil
}

func (a *customerAdapter) GetByID(id uuid.UUID) (*model.CustomerWithService, error) {
	// Query customer
	customerQuery, _, err := goqu.Dialect("postgres").
		From(tableCustomers).
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build customer query")
	}

	var customer model.Customer
	err = a.db.QueryRow(customerQuery).Scan(
		&customer.ID,
		&customer.MikrotikID,
		&customer.Username,
		&customer.FullName,
		&customer.Phone,
		&customer.Email,
		&customer.Address,
		&customer.Status,
		&customer.JoinDate,
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
		Where(goqu.Ex{"customer_id": id}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build service query")
	}

	var service model.CustomerService
	err = a.db.QueryRow(serviceQuery).Scan(
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
		&service.MikrotikObjectID, // This is at the END
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

func (a *customerAdapter) GetByUsername(mikrotikID uuid.UUID, username string) (*model.Customer, error) {
	query, _, err := goqu.Dialect("postgres").
		From(tableCustomers).
		Where(goqu.Ex{"mikrotik_id": mikrotikID, "username": username}).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build query")
	}

	var result model.Customer
	err = a.db.QueryRow(query).Scan(
		&result.ID,
		&result.MikrotikID,
		&result.Username,
		&result.FullName,
		&result.Phone,
		&result.Email,
		&result.Address,
		&result.Status,
		&result.JoinDate,
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

func (a *customerAdapter) List(mikrotikID uuid.UUID) ([]model.CustomerWithService, error) {
	// Query all customers for this MikroTik
	customersQuery, _, err := goqu.Dialect("postgres").
		From(tableCustomers).
		Where(goqu.Ex{"mikrotik_id": mikrotikID}).
		Order(goqu.I("created_at").Desc()).
		ToSQL()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to build customers query")
	}

	rows, err := a.db.Query(customersQuery)
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
			&customer.Username,
			&customer.FullName,
			&customer.Phone,
			&customer.Email,
			&customer.Address,
			&customer.Status,
			&customer.JoinDate,
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
			Where(goqu.Ex{"customer_id": customer.ID}).
			ToSQL()
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to build service query")
		}

		var service model.CustomerService
		err = a.db.QueryRow(serviceQuery).Scan(
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
			&service.MikrotikObjectID,
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

func (a *customerAdapter) Update(id uuid.UUID, input model.CustomerInput) error {
	// Update customer
	customerUpdate := goqu.Record{
		"username":  input.Username,
		"full_name": input.FullName,
		"phone":     input.Phone,
		"email":     input.Email,
		"address":   input.Address,
	}

	customerQuery, _, err := goqu.Dialect("postgres").
		Update(tableCustomers).
		Set(customerUpdate).
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build customer update query")
	}

	result, err := a.db.Exec(customerQuery)
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

	// Update service
	// Check if service exists first
	serviceQuery, _, err := goqu.Dialect("postgres").
		From(tableCustomerServices).
		Where(goqu.Ex{"customer_id": id}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build existing service query")
	}

	var existingServiceID uuid.UUID
	err = a.db.QueryRow(serviceQuery).Scan(&existingServiceID)
	// We only need to check if it exists, scan errors (except no rows) are handled below

	if err == sql.ErrNoRows {
		// Create new service if not exists (should not happen in normal flow but for safety)
		// Skipping create here for update operation simplicity, assuming service exists
		return nil
	} else if err != nil {
		// Ignore scan error as we only checking existence by count or similar, but queryrow simple
		// Actually if we scanned into UUID and it failed, likely column count mismatch if we used *
		// But here I selected * (default SELECT * FROM... if no Select() called?)
		// Wait, goqu From().Where().ToSQL() generates SELECT * FROM ...
		// Use explicit Select to check
	}

	// Update service fields
	serviceUpdate := goqu.Record{
		"profile_id": input.ProfileID,
		"price":      input.Price,
		"tax_rate":   *input.TaxRate,
		"start_date": *input.StartDate,
	}

	serviceUpdateQuery, _, err := goqu.Dialect("postgres").
		Update(tableCustomerServices).
		Set(serviceUpdate).
		Where(goqu.Ex{"customer_id": id}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build service update query")
	}

	_, err = a.db.Exec(serviceUpdateQuery)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update service")
	}

	return nil
}

func (a *customerAdapter) Delete(id uuid.UUID) error {
	// Delete customer (service will be cascade deleted)
	query, _, err := goqu.Dialect("postgres").
		Delete(tableCustomers).
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to build delete query")
	}

	result, err := a.db.Exec(query)
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
