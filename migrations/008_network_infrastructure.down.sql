-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_network_segments ON network_segments;
DROP TRIGGER IF EXISTS set_updated_at_cable_routes ON cable_routes;
DROP TABLE IF EXISTS network_segments;
DROP TABLE IF EXISTS cable_routes;
DROP TYPE IF EXISTS segment_status;
DROP TYPE IF EXISTS segment_type;

-- +goose StatementEnd