package controller

import (
	"context"
	"reflect"
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

func TestAddBook(t *testing.T) {
	t.Parallel()

	bookID := uuid.NewString()
	authorID := uuid.NewString()
	book := entity.Book{
		ID:        bookID,
		Name:      "Evgeny Onegin",
		AuthorIDs: []string{authorID},
	}

	testCases := []struct {
		name          string
		req           *library.AddBookRequest
		mockSetup     func(*lib.MockBooksUseCase)
		expectedError codes.Code
	}{
		{
			name: "Success",
			req:  &library.AddBookRequest{Name: book.Name, AuthorIds: book.AuthorIDs},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase) {
				mockBooksUseCase.EXPECT().
					RegisterBook(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, b entity.Book) (entity.Book, error) {
						if b.Name != book.Name || !reflect.DeepEqual(b.AuthorIDs, book.AuthorIDs) {
							return entity.Book{}, errors.New("unexpected book data")
						}
						b.ID = bookID
						return b, nil
					}).
					Times(1)
			},
			expectedError: codes.OK,
		},
		{
			name: "RegisterBookErrorBookAlreadyExists",
			req:  &library.AddBookRequest{Name: book.Name, AuthorIds: book.AuthorIDs},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase) {
				mockBooksUseCase.EXPECT().
					RegisterBook(gomock.Any(), gomock.Any()).
					Return(entity.Book{}, entity.ErrBookAlreadyExists).
					Times(1)
			},
			expectedError: codes.AlreadyExists,
		},

		{
			name:          "ValidationErrorEmptyName",
			req:           &library.AddBookRequest{Name: "", AuthorIds: []string{authorID}},
			mockSetup:     nil,
			expectedError: codes.InvalidArgument,
		},
		{
			name:          "ValidationErrorEmptyAuthorIDs",
			req:           &library.AddBookRequest{Name: book.Name, AuthorIds: []string{""}},
			mockSetup:     nil,
			expectedError: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl, controller, mockBooksUseCase, _ := setupTest(t)
			defer ctrl.Finish()
			if tc.mockSetup != nil {
				tc.mockSetup(mockBooksUseCase)
			}
			resp, err := controller.AddBook(context.Background(), tc.req)
			if tc.expectedError == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, book.ID, resp.GetBook().GetId())
				require.Equal(t, book.Name, resp.GetBook().GetName())
				require.Equal(t, book.AuthorIDs, resp.GetBook().GetAuthorId())
			} else {
				require.Error(t, err)
				require.Nil(t, resp)
				require.Equal(t, tc.expectedError, status.Code(err))
			}
		})
	}
}
