-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(255) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    user_id TEXT NOT NULL,
    is_deleted BOOLEAN DEFAULT false
    );
CREATE UNIQUE INDEX IF NOT EXISTS unique_original_url ON urls (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS unique_original_url;
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
