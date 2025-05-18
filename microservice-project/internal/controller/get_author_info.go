package controller

import (
	"context"

	"github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) GetAuthorInfo(ctx context.Context, req *library.GetAuthorInfoRequest) (*library.GetAuthorInfoResponse, error) {
	i.logger.Info("GetAuthorInfo request received", zap.String("authorID", req.GetId()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("GetAuthorInfo request validation failed", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	author, err := i.authorUseCase.GetAuthor(ctx, req.GetId())

	if err != nil {
		return nil, i.convertErr(err)
	}

	i.logger.Info("Author info received successfully",
		zap.String("authorID", req.GetId()))

	return &library.GetAuthorInfoResponse{
		Id:   author.ID,
		Name: author.Name,
	}, nil
}
