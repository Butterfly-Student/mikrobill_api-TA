-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_hotspot_vouchers ON hotspot_vouchers;
DROP TABLE IF EXISTS voucher_usage_logs;
DROP TABLE IF EXISTS voucher_batches;
DROP TABLE IF EXISTS hotspot_vouchers;
DROP TYPE IF EXISTS voucher_status;

-- +goose StatementEnd