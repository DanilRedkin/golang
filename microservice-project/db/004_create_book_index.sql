-- +goose Up
CREATE INDEX idx_book_name ON book (name);

-- +goose Down
DROP INDEX IF EXISTS idx_book_name;
