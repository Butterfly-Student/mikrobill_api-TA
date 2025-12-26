-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS set_updated_at_roles ON roles;
DROP TRIGGER IF EXISTS set_updated_at_users ON users;

DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;

-- +goose StatementEnd