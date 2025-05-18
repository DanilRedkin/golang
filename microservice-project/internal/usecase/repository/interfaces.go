package repository

//go:generate mockgen -source=interfaces.go -destination=repository_mock.go -package=repository

import (
	"context"
	"time"

	"github.com/project/library/internal/entity"
)

type (
	AuthorRepository interface {
		CreateAuthor(ctx context.Context, author entity.Author) (entity.Author, error)
		GetAuthor(ctx context.Context, authorID string) (entity.Author, error)
		ChangeAuthor(ctx context.Context, author entity.Author) error
	}

	BooksRepository interface {
		CreateBook(ctx context.Context, book entity.Book) (entity.Book, error)
		UpdateBook(ctx context.Context, book entity.Book) error
		GetBook(ctx context.Context, bookID string) (entity.Book, error)
		GetBooksByAuthor(ctx context.Context, authorID string) ([]entity.Book, error)
	}

	OutboxRepository interface {
		SendMessage(ctx context.Context, idempotencyKey string, kind OutboxKind, message []byte) error
		GetMessages(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]OutboxData, error)
		MarkAsProcessed(ctx context.Context, idempotencyKeys []string) error
	}

	Transactor interface {
		WithTx(ctx context.Context, function func(ctx context.Context) error) error
	}

	OutboxData struct {
		IdempotencyKey string
		Kind           OutboxKind
		RawData        []byte
	}
)

type OutboxKind int

const (
	OutboxKindUndefined OutboxKind = iota
	OutboxKindBook
	OutboxKindAuthor
)

func (o OutboxKind) String() string {
	switch o {
	case OutboxKindBook:
		return "book"
	case OutboxKindAuthor:
		return "author"
	default:
		return "undefined"
	}
}
