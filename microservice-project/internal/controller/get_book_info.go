package controller

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) GetBookInfo(ctx context.Context, req *library.GetBookInfoRequest) (*library.GetBookInfoResponse, error) {
	i.logger.Info("GetBookInfo request received", zap.String("bookID", req.GetId()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("GetBookInfo request validation failed", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	book, err := i.booksUseCase.GetBook(ctx, req.GetId())
	if err != nil {
		return nil, i.convertErr(err)
	}

	i.logger.Info("Book info received successfully", zap.String("bookID", req.GetId()))

	return &library.GetBookInfoResponse{
		Book: &library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		},
	}, nil
}
