-- +goose Up
CREATE INDEX idx_author_name ON author (name);

-- +goose Down
DROP INDEX IF EXISTS idx_author_name;
