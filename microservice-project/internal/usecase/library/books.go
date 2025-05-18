package library

import (
	"context"
	"encoding/json"

	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

func (l *Impl) RegisterBook(ctx context.Context, book entity.Book) (entity.Book, error) {
	var createdBook entity.Book

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var txErr error

		createdBook, txErr = l.booksRepository.CreateBook(ctx, book)
		if txErr != nil {
			l.logger.Error("Failed to register book",
				zap.Any("book", book),
				zap.Error(txErr),
			)
			return txErr
		}

		serialized, txErr := json.Marshal(createdBook)
		if txErr != nil {
			l.logger.Error("Failed to serialize created book",
				zap.String("bookID", createdBook.ID),
				zap.Error(txErr),
			)
			return txErr
		}

		key := repository.OutboxKindBook.String() + "_" + createdBook.ID
		txErr = l.outboxRepository.SendMessage(ctx, key, repository.OutboxKindBook, serialized)
		if txErr != nil {
			l.logger.Error("Failed to send outbox message for book",
				zap.String("outboxKey", key),
				zap.Error(txErr),
			)
			return txErr
		}

		return nil
	})

	if err != nil {
		l.logger.Error("Transaction failed for RegisterBook",
			zap.Any("book", book),
			zap.Error(err),
		)
		return entity.Book{}, err
	}

	return createdBook, nil
}

func (l *Impl) GetBook(ctx context.Context, bookID string) (entity.Book, error) {
	book, err := l.booksRepository.GetBook(ctx, bookID)
	if err != nil {
		l.logger.Error("Failed to get book",
			zap.String("bookID", bookID),
			zap.Error(err))
		return entity.Book{}, err
	}

	return book, nil
}

func (l *Impl) UpdateBook(ctx context.Context, bookID string, name string, authorIDs []string) error {
	updatedBook := entity.Book{
		ID:        bookID,
		Name:      name,
		AuthorIDs: authorIDs,
	}

	if err := l.booksRepository.UpdateBook(ctx, updatedBook); err != nil {
		l.logger.Error("Failed to update book in repository",
			zap.String("bookID", bookID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (l *Impl) GetAuthorBooks(ctx context.Context, authorID string) ([]entity.Book, error) {
	books, err := l.booksRepository.GetBooksByAuthor(ctx, authorID)
	if err != nil {
		l.logger.Error("Failed to get books by author",
			zap.String("authorID", authorID),
			zap.Error(err))
		return nil, err
	}

	return books, nil
}
