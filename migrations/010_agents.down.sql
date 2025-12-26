-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_agents ON agents;
DROP TABLE IF EXISTS agent_transactions;
DROP TABLE IF EXISTS agent_balances;
DROP TABLE IF EXISTS agents;
DROP TYPE IF EXISTS voucher_status;
DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS agent_transaction_type;
DROP TYPE IF EXISTS agent_status;

-- +goose StatementEnd