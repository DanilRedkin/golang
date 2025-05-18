package controller

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/entity"
	lib "github.com/project/library/mocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetAuthorBooks(t *testing.T) {
	t.Parallel()
	var (
		book1id  = uuid.NewString()
		book2id  = uuid.NewString()
		authorID = uuid.NewString()
	)

	book1 := entity.Book{
		ID:        book1id,
		Name:      "Evgeny Onegin",
		AuthorIDs: []string{authorID},
	}

	book2 := entity.Book{
		ID:        book2id,
		Name:      "Evgeny Onegin 2",
		AuthorIDs: []string{authorID},
	}

	books := []entity.Book{book1, book2}

	testCases := []struct {
		name          string
		req           *library.GetAuthorBooksRequest
		mockSetup     func(*lib.MockBooksUseCase, *lib.MockLibrary_GetAuthorBooksServer)
		expectedError codes.Code
	}{
		{
			name: "Success",
			req: &library.GetAuthorBooksRequest{
				AuthorId: authorID,
			},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase, mockServer *lib.MockLibrary_GetAuthorBooksServer) {
				mockServer.EXPECT().Context().Return(context.Background()).Times(1)
				mockBooksUseCase.EXPECT().
					GetAuthorBooks(gomock.Any(), authorID).
					Return(books, nil).
					Times(1)
				mockServer.EXPECT().Send(gomock.Any()).Times(len(books)).Do(func(book *library.Book) {
					require.Contains(t, []string{book1id, book2id}, book.GetId())
					require.Contains(t, []string{book1.Name, book2.Name}, book.GetName())
					require.ElementsMatch(t, []string{authorID}, book.GetAuthorId())
				})
			},
			expectedError: codes.OK,
		},
		{
			name: "NoBooksFound",
			req: &library.GetAuthorBooksRequest{
				AuthorId: authorID,
			},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase, mockServer *lib.MockLibrary_GetAuthorBooksServer) {
				mockServer.EXPECT().Context().Return(context.Background()).Times(1)
				mockBooksUseCase.EXPECT().
					GetAuthorBooks(gomock.Any(), authorID).
					Return(nil, entity.ErrBookNotFound).
					Times(1)
			},
			expectedError: codes.NotFound,
		},
		{
			name: "AuthorNotFound",
			req: &library.GetAuthorBooksRequest{
				AuthorId: authorID,
			},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase, mockServer *lib.MockLibrary_GetAuthorBooksServer) {
				mockServer.EXPECT().Context().Return(context.Background()).Times(1)
				mockBooksUseCase.EXPECT().
					GetAuthorBooks(gomock.Any(), authorID).
					Return(nil, entity.ErrAuthorNotFound).
					Times(1)
			},
			expectedError: codes.NotFound,
		},
		{
			name: "InternalErrorGetBooks",
			req: &library.GetAuthorBooksRequest{
				AuthorId: authorID,
			},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase, mockServer *lib.MockLibrary_GetAuthorBooksServer) {
				mockServer.EXPECT().Context().Return(context.Background()).Times(1)
				mockBooksUseCase.EXPECT().
					GetAuthorBooks(gomock.Any(), authorID).
					Return(nil, errors.New("internal error")).
					Times(1)
			},
			expectedError: codes.Internal,
		},
		{
			name: "InternalServerErrorSendFailure",
			req: &library.GetAuthorBooksRequest{
				AuthorId: authorID,
			},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase, mockServer *lib.MockLibrary_GetAuthorBooksServer) {
				mockServer.EXPECT().Context().Return(context.Background()).Times(1)
				mockBooksUseCase.EXPECT().
					GetAuthorBooks(gomock.Any(), authorID).
					Return(books, nil).
					Times(1)
				mockServer.EXPECT().Send(gomock.Any()).Return(errors.New("send error")).Times(1)
			},
			expectedError: codes.Internal,
		},
		{
			name: "ValidationErrorEmptyAuthorID",
			req: &library.GetAuthorBooksRequest{
				AuthorId: "",
			},
			mockSetup:     nil,
			expectedError: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBooksUseCase := lib.NewMockBooksUseCase(ctrl)
			mockServer := lib.NewMockLibrary_GetAuthorBooksServer(ctrl)
			controller := &Implementation{
				booksUseCase:  mockBooksUseCase,
				logger:        logger,
				authorUseCase: nil,
			}

			if tc.mockSetup != nil {
				tc.mockSetup(mockBooksUseCase, mockServer)
			}

			err := controller.GetAuthorBooks(tc.req, mockServer)

			if tc.expectedError == codes.OK {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.expectedError, status.Code(err))
			}
		})
	}
}
