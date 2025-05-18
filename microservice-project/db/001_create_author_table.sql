-- +goose Up
CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE author
(
    id         UUID PRIMARY KEY     DEFAULT uuid_generate_v4(),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose StatementBegin
CREATE
OR REPLACE FUNCTION update_author_timestamp() RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at
= now();
RETURN NEW;
END;
$$
LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trigger_update_author_timestamp
    BEFORE UPDATE
    ON author
    FOR EACH ROW
    EXECUTE FUNCTION update_author_timestamp();

-- +goose Down
DROP TABLE IF EXISTS author;
