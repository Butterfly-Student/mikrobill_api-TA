-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_odp_connections ON odp_connections;
DROP TRIGGER IF EXISTS set_updated_at_odps ON odps;
DROP TABLE IF EXISTS odp_connections;
DROP TABLE IF EXISTS odps;
DROP TYPE IF EXISTS cable_capacity;
DROP TYPE IF EXISTS connection_type;
DROP TYPE IF EXISTS odp_status;

-- +goose StatementEnd