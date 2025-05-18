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

func TestUpdateBook(t *testing.T) {
	t.Parallel()
	bookID := uuid.NewString()
	authorID := uuid.NewString()
	newTitle := "New Title"

	testCases := []struct {
		name          string
		req           *library.UpdateBookRequest
		mockSetup     func(*lib.MockBooksUseCase)
		expectedError codes.Code
	}{
		{
			name: "Success",
			req: &library.UpdateBookRequest{
				Id:        bookID,
				Name:      newTitle,
				AuthorIds: []string{authorID},
			},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase) {
				mockBooksUseCase.EXPECT().
					UpdateBook(gomock.Any(), bookID, newTitle, []string{authorID}).
					Return(nil).
					Times(1)
			},
			expectedError: codes.OK,
		},
		{
			name: "BookUpdateFailure",
			req: &library.UpdateBookRequest{
				Id:        bookID,
				Name:      newTitle,
				AuthorIds: []string{authorID},
			},
			mockSetup: func(mockBooksUseCase *lib.MockBooksUseCase) {
				mockBooksUseCase.EXPECT().
					UpdateBook(gomock.Any(), bookID, newTitle, []string{authorID}).
					Return(entity.ErrBookNotFound).
					Times(1)
			},
			expectedError: codes.NotFound,
		},
		{
			name: "ValidationErrorEmptyID",
			req: &library.UpdateBookRequest{
				Id:        "",
				Name:      newTitle,
				AuthorIds: []string{authorID},
			},
			mockSetup:     nil,
			expectedError: codes.InvalidArgument,
		},
		{
			name: "ValidationErrorEmptyName",
			req: &library.UpdateBookRequest{
				Id:        bookID,
				Name:      "",
				AuthorIds: []string{authorID},
			},
			mockSetup:     nil,
			expectedError: codes.InvalidArgument,
		},
		{
			name: "ValidationErrorEmptyAuthorIDs",
			req: &library.UpdateBookRequest{
				Id:        bookID,
				Name:      newTitle,
				AuthorIds: []string{""},
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
			resp, err := controller.UpdateBook(context.Background(), tc.req)
			if tc.expectedError == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, resp)
			} else {
				require.Error(t, err)
				require.Nil(t, resp)
				require.Equal(t, tc.expectedError, status.Code(err))
			}
		})
	}
}
