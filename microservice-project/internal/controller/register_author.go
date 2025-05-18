package controller

import (
	"context"

	"github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) RegisterAuthor(ctx context.Context, req *library.RegisterAuthorRequest) (*library.RegisterAuthorResponse, error) {
	i.logger.Info("RegisterAuthor request received", zap.String("name", req.GetName()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("RegisterAuthor request validation failed", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.RegisterAuthor(ctx, req.GetName())
	if err != nil {
		return nil, i.convertErr(err)
	}

	i.logger.Info("Author added successfully",
		zap.String("authorID", author.ID),
		zap.String("name", author.Name))

	return &library.RegisterAuthorResponse{
		Id: author.ID,
	}, nil
}
