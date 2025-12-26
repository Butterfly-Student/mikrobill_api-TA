-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_expenses ON expenses;
DROP TRIGGER IF EXISTS set_updated_at_collector_payments ON collector_payments;
DROP TRIGGER IF EXISTS set_updated_at_collectors ON collectors;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS collector_payments;
DROP TABLE IF EXISTS collectors;
DROP TYPE IF EXISTS collector_payment_status;
DROP TYPE IF EXISTS payment_method;
DROP TYPE IF EXISTS collector_status;

-- +goose StatementEnd