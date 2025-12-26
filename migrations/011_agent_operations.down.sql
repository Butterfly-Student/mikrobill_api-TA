-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS agent_notifications;
DROP TABLE IF EXISTS agent_payments;
DROP TABLE IF EXISTS agent_monthly_payments;
DROP TABLE IF EXISTS agent_balance_requests;
DROP TABLE IF EXISTS agent_voucher_sales;
DROP TYPE IF EXISTS notification_type;
DROP TYPE IF EXISTS request_status;
