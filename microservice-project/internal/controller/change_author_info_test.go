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

func TestChangeAuthorInfo(t *testing.T) {
	t.Parallel()
	authorID := uuid.NewString()
	newName := "New Name"

	testCases := []struct {
		name          string
		req           *library.ChangeAuthorInfoRequest
		mockSetup     func(*lib.MockAuthorUseCase)
		expectedError codes.Code
	}{
		{
			name: "Success",
			req: &library.ChangeAuthorInfoRequest{
				Id:   authorID,
				Name: newName,
			},
			mockSetup: func(mockAuthorUseCase *lib.MockAuthorUseCase) {
				mockAuthorUseCase.EXPECT().
					ChangeAuthor(gomock.Any(), authorID, newName).
					Return(nil).
					Times(1)
			},
			expectedError: codes.OK,
		},
		{
			name: "Failure_AuthorUpdateError",
			req: &library.ChangeAuthorInfoRequest{
				Id:   authorID,
				Name: newName,
			},
			mockSetup: func(mockAuthorUseCase *lib.MockAuthorUseCase) {
				mockAuthorUseCase.EXPECT().
					ChangeAuthor(gomock.Any(), authorID, newName).
					Return(entity.ErrAuthorNotFound).
					Times(1)
			},
			expectedError: codes.NotFound,
		},
		{
			name: "ValidationErrorEmptyID",
			req: &library.ChangeAuthorInfoRequest{
				Id:   "",
				Name: newName,
			},
			mockSetup:     nil,
			expectedError: codes.InvalidArgument,
		},
		{
			name: "ValidationErrorEmptyName",
			req: &library.ChangeAuthorInfoRequest{
				Id:   authorID,
				Name: "",
			},
			mockSetup:     nil,
			expectedError: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl, controller, _, mockAuthorUseCase := setupTest(t)
			defer ctrl.Finish()
			if tc.mockSetup != nil {
				tc.mockSetup(mockAuthorUseCase)
			}
			resp, err := controller.ChangeAuthorInfo(context.Background(), tc.req)
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
