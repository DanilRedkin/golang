package controller

import (
	"context"

	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/entity"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *Implementation) AddBook(ctx context.Context, req *library.AddBookRequest) (*library.AddBookResponse, error) {
	i.logger.Info("AddBook request received",
		zap.String("name", req.GetName()),
		zap.Strings("authorIDs", req.GetAuthorIds()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("AddBook request validation failed", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book := entity.Book{
		ID:        uuid.New().String(),
		Name:      req.GetName(),
		AuthorIDs: req.GetAuthorIds(),
	}

	createdBook, err := i.booksUseCase.RegisterBook(ctx, book)
	if err != nil {
		i.logger.Error("Failed to register book", zap.Error(err))
		return nil, i.convertErr(err)
	}

	bookResponse := &library.AddBookResponse{
		Book: &library.Book{
			Id:        createdBook.ID,
			Name:      createdBook.Name,
			AuthorId:  createdBook.AuthorIDs,
			CreatedAt: timestamppb.New(createdBook.CreatedAt),
			UpdatedAt: timestamppb.New(createdBook.UpdatedAt),
		},
	}

	i.logger.Info("Book added successfully",
		zap.String("bookID", createdBook.ID),
		zap.String("name", createdBook.Name))

	return bookResponse, nil
}
