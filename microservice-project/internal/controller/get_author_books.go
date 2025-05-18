package controller

import (
	"github.com/project/library/generated/api/library"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *Implementation) GetAuthorBooks(req *library.GetAuthorBooksRequest, server library.Library_GetAuthorBooksServer) error {
	i.logger.Info("GetAuthorBooks request received", zap.String("authorID", req.GetAuthorId()))

	if err := req.ValidateAll(); err != nil {
		i.logger.Error("GetAuthorBooks request validation failed", zap.Error(err))
		return status.Error(codes.InvalidArgument, err.Error())
	}

	books, err := i.booksUseCase.GetAuthorBooks(
		server.Context(),
		req.GetAuthorId())

	if err != nil {
		return i.convertErr(err)
	}

	for _, book := range books {
		if err := server.Send(&library.Book{
			Id:        book.ID,
			Name:      book.Name,
			AuthorId:  book.AuthorIDs,
			CreatedAt: timestamppb.New(book.CreatedAt),
			UpdatedAt: timestamppb.New(book.UpdatedAt),
		}); err != nil {
			i.logger.Error("Failed to send book",
				zap.String("bookID", book.ID),
				zap.Error(err))

			return status.Errorf(codes.Internal, "failed to send book: %v", err)
		}
	}

	i.logger.Info("Books sent successfully", zap.String("authorID", req.GetAuthorId()))
	return nil
}
