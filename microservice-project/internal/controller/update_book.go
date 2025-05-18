package controller

import (
	"context"

	"github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) UpdateBook(ctx context.Context, req *library.UpdateBookRequest) (*library.UpdateBookResponse, error) {
	i.logger.Info("UpdateBook request received",
		zap.String("bookID", req.GetId()),
		zap.String("name", req.GetName()),
		zap.Strings("authorIDs", req.GetAuthorIds()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("UpdateBook request validation failed", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.booksUseCase.UpdateBook(ctx, req.GetId(), req.GetName(), req.GetAuthorIds())

	if err != nil {
		return nil, i.convertErr(err)
	}

	i.logger.Info("Book updated successfully", zap.String("bookID", req.GetId()))

	return &library.UpdateBookResponse{}, nil
}
