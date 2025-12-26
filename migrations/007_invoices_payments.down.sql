-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_pgt ON payment_gateway_transactions;
DROP TRIGGER IF EXISTS set_updated_at_invoices ON invoices;
DROP TABLE IF EXISTS payment_gateway_transactions;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoices;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS invoice_type;
DROP TYPE IF EXISTS invoice_status;

-- +goose StatementEnd