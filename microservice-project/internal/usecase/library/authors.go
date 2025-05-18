package library

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/repository"
	"go.uber.org/zap"
)

func (l *Impl) RegisterAuthor(ctx context.Context, authorName string) (entity.Author, error) {
	var createdAuthor entity.Author

	err := l.transactor.WithTx(ctx, func(ctx context.Context) error {
		var txErr error

		createdAuthor, txErr = l.authorRepository.CreateAuthor(ctx,
			entity.Author{
				ID:   uuid.New().String(),
				Name: authorName,
			})

		if txErr != nil {
			l.logger.Error("Failed to register author",
				zap.String("authorName", authorName),
				zap.Error(txErr),
			)
			return txErr
		}

		serialized, txErr := json.Marshal(createdAuthor)
		if txErr != nil {
			l.logger.Error("Failed to serialize created author",
				zap.String("authorID", createdAuthor.ID),
				zap.Error(txErr),
			)
			return txErr
		}

		key := repository.OutboxKindAuthor.String() + "_" + createdAuthor.ID
		txErr = l.outboxRepository.SendMessage(ctx, key, repository.OutboxKindAuthor, serialized)
		if txErr != nil {
			l.logger.Error("Failed to send outbox message",
				zap.String("outboxKey", key),
				zap.Error(txErr),
			)
			return txErr
		}

		return nil
	})

	if err != nil {
		l.logger.Error("Transaction failed for RegisterAuthor",
			zap.String("authorName", authorName),
			zap.Error(err),
		)
		return entity.Author{}, err
	}

	return createdAuthor, nil
}

func (l *Impl) GetAuthor(ctx context.Context, authorID string) (entity.Author, error) {
	author, err := l.authorRepository.GetAuthor(ctx, authorID)

	if err != nil {
		l.logger.Error("Failed to get author", zap.String("authorID", authorID), zap.Error(err))
		return entity.Author{}, err
	}
	return author, nil
}

func (l *Impl) ChangeAuthor(ctx context.Context, authorID string, newAuthorName string) error {
	author := entity.Author{
		ID:   authorID,
		Name: newAuthorName,
	}

	if err := l.authorRepository.ChangeAuthor(ctx, author); err != nil {
		l.logger.Error("Failed to change author", zap.String("authorID", authorID), zap.Error(err))
		return err
	}

	return nil
}
