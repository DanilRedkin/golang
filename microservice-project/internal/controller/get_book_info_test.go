package controller

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/project/library/generated/api/library"
	"github.com/project/library/internal/entity"
	lib "github.com/project/library/mocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetBookInfo(t *testing.T) {
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
		req           *library.GetBookInfoRequest
		mockSetup     func(*lib.MockBooksUseCase)
		expectedError codes.Code
	}{
		{
			name: "Success",
			req: &library.GetBookInfoRequest{
				Id: bookID,
			},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase) {
				mockBooksUseCase.EXPECT().
					GetBook(gomock.Any(), bookID).
					Return(book, nil).
					Times(1)
			},
			expectedError: codes.OK,
		},
		{
			name: "BookNotFound",
			req: &library.GetBookInfoRequest{
				Id: bookID,
			},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase) {
				mockBooksUseCase.EXPECT().
					GetBook(gomock.Any(), bookID).
					Return(entity.Book{}, entity.ErrBookNotFound).
					Times(1)
			},
			expectedError: codes.NotFound,
		},
		{
			name: "ValidationErrorEmptyID",
			req: &library.GetBookInfoRequest{
				Id: "",
			},
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
			resp, err := controller.GetBookInfo(context.Background(), tc.req)
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
