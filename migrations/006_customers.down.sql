-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_customers ON customers;
DROP TABLE IF EXISTS customers;
DROP TYPE IF EXISTS service_type;
DROP TYPE IF EXISTS cable_status;
DROP TYPE IF EXISTS customer_status;

-- +goose StatementEnd