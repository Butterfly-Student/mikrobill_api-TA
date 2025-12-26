-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_onu_devices ON onu_devices;
DROP TABLE IF EXISTS onu_devices;
DROP TABLE IF EXISTS cable_maintenance_logs;
DROP TYPE IF EXISTS device_status;
DROP TYPE IF EXISTS maintenance_type;

-- +goose StatementEnd