package library

//go:generate mockgen -source=interfaces.go -destination=library_mock.go -package=library

import (
	"context"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

type (
	AuthorUseCase interface {
		RegisterAuthor(ctx context.Context, authorName string) (entity.Author, error)
		GetAuthor(ctx context.Context, authorID string) (entity.Author, error)
		ChangeAuthor(ctx context.Context, authorID string, authorName string) error
	}

	BooksUseCase interface {
		UpdateBook(ctx context.Context, bookID string, name string, authorIDs []string) error
		RegisterBook(ctx context.Context, book entity.Book) (entity.Book, error)
		GetBook(ctx context.Context, bookID string) (entity.Book, error)
		GetAuthorBooks(ctx context.Context, authorID string) ([]entity.Book, error)
	}
)

var _ AuthorUseCase = (*Impl)(nil)
var _ BooksUseCase = (*Impl)(nil)

type Impl struct {
	logger           *zap.Logger
	authorRepository repository.AuthorRepository
	booksRepository  repository.BooksRepository
	outboxRepository repository.OutboxRepository
	transactor       repository.Transactor
}

func New(
	logger *zap.Logger,
	authorRepository repository.AuthorRepository,
	booksRepository repository.BooksRepository,
	outboxRepository repository.OutboxRepository,
	transactor repository.Transactor,
) *Impl {
	return &Impl{
		logger:           logger,
		authorRepository: authorRepository,
		booksRepository:  booksRepository,
		outboxRepository: outboxRepository,
		transactor:       transactor,
	}
}
