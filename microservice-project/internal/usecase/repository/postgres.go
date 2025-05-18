package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/project/library/internal/entity"
)

var (
	_ BooksRepository  = (*postgresRepository)(nil)
	_ AuthorRepository = (*postgresRepository)(nil)
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{db: db}
}

func (p *postgresRepository) CreateBook(ctx context.Context, book entity.Book) (resBook entity.Book, txErr error) {
	var (
		tx  pgx.Tx
		err error
	)

	if tx, err = extractTx(ctx); err != nil {
		tx, err = p.db.Begin(ctx)

		if err != nil {
			return entity.Book{}, err
		}

		defer commitOrRollbackTx(ctx, tx, &txErr)
	}

	if err != nil {
		return entity.Book{}, err
	}

	if err := p.validateAuthorsExist(ctx, tx, book.AuthorIDs); err != nil {
		return entity.Book{}, err
	}

	const queryBook = `
		INSERT INTO book (id, name, created_at, updated_at)
		VALUES ($1, $2, now(), now())
		RETURNING created_at, updated_at
	`

	result := entity.Book{
		ID:        book.ID,
		Name:      book.Name,
		AuthorIDs: book.AuthorIDs,
	}

	if err := tx.QueryRow(ctx, queryBook, book.ID, book.Name).Scan(&result.CreatedAt, &result.UpdatedAt); err != nil {
		return entity.Book{}, fmt.Errorf("inserting book: %w", err)
	}

	if err := p.insertAuthorBooks(ctx, tx, book.ID, book.AuthorIDs); err != nil {
		return entity.Book{}, fmt.Errorf("inserting author-book relations: %w", err)
	}

	return result, nil
}

func (p *postgresRepository) UpdateBook(ctx context.Context, book entity.Book) (err error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() {
		if rbErr := p.rollbackTx(ctx, tx, err); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			if err != nil {
				err = fmt.Errorf("%v; rollback error: %w", err, rbErr)
			} else {
				err = fmt.Errorf("rollback error: %w", rbErr)
			}
		}
	}()

	if err := p.validateAuthorsExist(ctx, tx, book.AuthorIDs); err != nil {
		return fmt.Errorf("validating authors: %w", err)
	}

	if err := p.checkBookExists(ctx, tx, book.ID); err != nil {
		return fmt.Errorf("checking book existence: %w", err)
	}

	if err := p.updateBookDetails(ctx, tx, book.ID, book.Name); err != nil {
		return fmt.Errorf("updating book details: %w", err)
	}

	if err := p.updateBookAuthors(ctx, tx, book.ID, book.AuthorIDs); err != nil {
		return fmt.Errorf("updating book authors: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (p *postgresRepository) GetBook(ctx context.Context, bookID string) (entity.Book, error) {
	const queryBook = `
		SELECT id, name, created_at, updated_at
		FROM book
		WHERE id = $1
	`

	var book entity.Book
	if err := p.db.QueryRow(ctx, queryBook, bookID).Scan(
		&book.ID,
		&book.Name,
		&book.CreatedAt,
		&book.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Book{}, entity.ErrBookNotFound
		}
		return entity.Book{}, fmt.Errorf("querying book: %w", err)
	}

	authors, err := p.getBookAuthors(ctx, bookID)
	if err != nil {
		return entity.Book{}, fmt.Errorf("getting book authors: %w", err)
	}
	book.AuthorIDs = authors

	return book, nil
}

func (p *postgresRepository) GetBooksByAuthor(ctx context.Context, authorID string) ([]entity.Book, error) {
	const query = `
		SELECT b.id, b.name, b.created_at, b.updated_at,
			(SELECT array_agg(ab2.author_id) FROM author_book ab2 WHERE ab2.book_id = b.id) AS author_ids
		FROM book b
		JOIN author_book ab ON ab.book_id = b.id
		WHERE ab.author_id = $1
		GROUP BY b.id, b.name, b.created_at, b.updated_at
	`

	rows, err := p.db.Query(ctx, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("querying books by author: %w", err)
	}
	defer rows.Close()

	var books []entity.Book
	for rows.Next() {
		var book entity.Book
		if err := rows.Scan(
			&book.ID,
			&book.Name,
			&book.CreatedAt,
			&book.UpdatedAt,
			&book.AuthorIDs,
		); err != nil {
			return nil, fmt.Errorf("scanning book row: %w", err)
		}
		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("processing rows: %w", err)
	}

	return books, nil
}

func (p *postgresRepository) CreateAuthor(ctx context.Context, author entity.Author) (resAuthor entity.Author, txErr error) {
	var (
		tx  pgx.Tx
		err error
	)
	if tx, err = extractTx(ctx); err != nil {
		tx, err = p.db.Begin(ctx)
		if err != nil {
			return entity.Author{}, fmt.Errorf("starting transaction: %w", err)
		}
		defer commitOrRollbackTx(ctx, tx, &txErr)
	}
	if err != nil {
		return entity.Author{}, err
	}

	const query = `
		INSERT INTO author (id, name, created_at, updated_at)
		VALUES ($1, $2, now(), now())
		RETURNING created_at, updated_at
	`

	if err := tx.QueryRow(ctx, query, author.ID, author.Name).
		Scan(&author.CreatedAt, &author.UpdatedAt); err != nil {
		return entity.Author{}, fmt.Errorf("creating author: %w", err)
	}

	return author, nil
}

func (p *postgresRepository) GetAuthor(ctx context.Context, authorID string) (entity.Author, error) {
	const query = `
		SELECT id, name, created_at, updated_at
		FROM author
		WHERE id = $1
	`

	var author entity.Author
	if err := p.db.QueryRow(ctx, query, authorID).Scan(
		&author.ID,
		&author.Name,
		&author.CreatedAt,
		&author.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Author{}, entity.ErrAuthorNotFound
		}
		return entity.Author{}, fmt.Errorf("getting author: %w", err)
	}

	return author, nil
}

func (p *postgresRepository) ChangeAuthor(ctx context.Context, author entity.Author) error {
	const query = `
		UPDATE author
		SET name = $1, updated_at = now()
		WHERE id = $2
	`

	tag, err := p.db.Exec(ctx, query, author.Name, author.ID)
	if err != nil {
		return fmt.Errorf("updating author: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return entity.ErrAuthorNotFound
	}

	return nil
}

func (p *postgresRepository) validateAuthorsExist(ctx context.Context, tx pgx.Tx, authorIDs []string) error {
	const query = `
SELECT 1
FROM author
WHERE id = $1
`

	for _, authorID := range authorIDs {
		var exists int
		if err := tx.QueryRow(ctx, query, authorID).Scan(&exists); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: author ID %s", entity.ErrAuthorNotFound, authorID)
			}
			return fmt.Errorf("checking author existence: %w", err)
		}
	}
	return nil
}

func (p *postgresRepository) rollbackTx(ctx context.Context, tx pgx.Tx, originalErr error) error {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		if originalErr != nil {
			return fmt.Errorf("%v; rollback error: %v", originalErr, err)
		}
		return err
	}
	return nil
}

func (p *postgresRepository) insertAuthorBooks(ctx context.Context, tx pgx.Tx, bookID string, authorIDs []string) error {
	rows := make([][]interface{}, len(authorIDs))
	for i, aid := range authorIDs {
		rows[i] = []interface{}{aid, bookID}
	}

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"author_book"},
		[]string{"author_id", "book_id"},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (p *postgresRepository) checkBookExists(ctx context.Context, tx pgx.Tx, bookID string) error {
	const query = `SELECT 1 FROM book WHERE id = $1`

	var exists int
	if err := tx.QueryRow(ctx, query, bookID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.ErrBookNotFound
		}
		return err
	}
	return nil
}

func (p *postgresRepository) updateBookDetails(ctx context.Context, tx pgx.Tx, bookID, name string) error {
	const query = `UPDATE book SET name = $1, updated_at = now() WHERE id = $2`

	tag, err := tx.Exec(ctx, query, name, bookID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return entity.ErrBookNotFound
	}
	return nil
}

func (p *postgresRepository) updateBookAuthors(ctx context.Context, tx pgx.Tx, bookID string, authorIDs []string) error {
	const deleteQuery = `DELETE FROM author_book WHERE book_id = $1`
	if _, err := tx.Exec(ctx, deleteQuery, bookID); err != nil {
		return fmt.Errorf("deleting old author relations: %w", err)
	}

	return p.insertAuthorBooks(ctx, tx, bookID, authorIDs)
}

func (p *postgresRepository) getBookAuthors(ctx context.Context, bookID string) ([]string, error) {
	const query = `SELECT author_id FROM author_book WHERE book_id = $1`

	rows, err := p.db.Query(ctx, query, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []string
	for rows.Next() {
		var authorID string
		if err := rows.Scan(&authorID); err != nil {
			return nil, err
		}
		authors = append(authors, authorID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return authors, nil
}

func commitOrRollbackTx(ctx context.Context, tx pgx.Tx, txErr *error) {
	if *txErr != nil {
		_ = tx.Rollback(ctx)
	} else {
		*txErr = tx.Commit(ctx)
	}
}
