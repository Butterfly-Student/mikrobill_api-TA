-- +goose Down
-- +goose StatementBegin
ALTER TABLE customers DROP COLUMN IF EXISTS interface;
-- +goose StatementEnd
