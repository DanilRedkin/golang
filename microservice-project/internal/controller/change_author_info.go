package controller

import (
	"context"

	"github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) ChangeAuthorInfo(ctx context.Context, req *library.ChangeAuthorInfoRequest) (*library.ChangeAuthorInfoResponse, error) {
	i.logger.Info("ChangeAuthorInfo request received",
		zap.String("authorID", req.GetId()),
		zap.String("name", req.GetName()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("ChangeAuthorInfo request validation failed", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.authorUseCase.ChangeAuthor(
		ctx,
		req.GetId(),
		req.GetName())

	if err != nil {
		return nil, i.convertErr(err)
	}

	i.logger.Info("Author info changed successfully",
		zap.String("authorID", req.GetId()))

	return &library.ChangeAuthorInfoResponse{}, nil
}
