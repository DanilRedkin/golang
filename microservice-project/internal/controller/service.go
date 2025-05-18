package controller

import (
	generated "github.com/project/library/generated/api/library"
	"github.com/project/library/internal/usecase/library"
	"go.uber.org/zap"
)

var _ generated.LibraryServer = (*Implementation)(nil)

type Implementation struct {
	logger        *zap.Logger
	booksUseCase  library.BooksUseCase
	authorUseCase library.AuthorUseCase
}

func New(
	logger *zap.Logger,
	booksUseCase library.BooksUseCase,
	authorUseCase library.AuthorUseCase,
) *Implementation {
	return &Implementation{
		logger:        logger,
		booksUseCase:  booksUseCase,
		authorUseCase: authorUseCase,
	}
}
