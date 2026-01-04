-- +goose Up
-- +goose StatementBegin
ALTER TABLE customers ADD COLUMN IF NOT EXISTS interface VARCHAR(255);
-- +goose StatementEnd
