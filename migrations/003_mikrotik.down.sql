-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_mikrotik ON mikrotik;
DROP TABLE IF EXISTS mikrotik;
DROP TYPE IF EXISTS mikrotik_status;

-- +goose StatementEnd