-- +goose Up
CREATE TABLE author_book
(
    author_id UUID NOT NULL,
    book_id   UUID NOT NULL,
    PRIMARY KEY (author_id, book_id),
    FOREIGN KEY (author_id) REFERENCES author (id) ON DELETE CASCADE,
    FOREIGN KEY (book_id) REFERENCES book (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS author_book;
