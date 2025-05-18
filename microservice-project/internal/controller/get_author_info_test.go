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

func TestGetAuthorInfo(t *testing.T) {
	t.Parallel()
	authorID := uuid.NewString()
	authorName := "John Doe"

	testCases := []struct {
		name          string
		req           *library.GetAuthorInfoRequest
		mockSetup     func(*lib.MockAuthorUseCase)
		expectedError codes.Code
	}{
		{
			name: "Success",
			req: &library.GetAuthorInfoRequest{
				Id: authorID,
			},
			mockSetup: func(mockAuthorUseCase *lib.MockAuthorUseCase) {
				mockAuthorUseCase.EXPECT().
					GetAuthor(gomock.Any(), authorID).
					Return(entity.Author{ID: authorID, Name: authorName}, nil).
					Times(1)
			},
			expectedError: codes.OK,
		},
		{
			name: "AuthorNotFound",
			req: &library.GetAuthorInfoRequest{
				Id: authorID,
			},
			mockSetup: func(mockAuthorUseCase *lib.MockAuthorUseCase) {
				mockAuthorUseCase.EXPECT().
					GetAuthor(gomock.Any(), authorID).
					Return(entity.Author{}, entity.ErrAuthorNotFound).
					Times(1)
			},
			expectedError: codes.NotFound,
		},
		{
			name: "ValidationErrorEmptyID",
			req: &library.GetAuthorInfoRequest{
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
			ctrl, controller, _, mockAuthorUseCase := setupTest(t)
			defer ctrl.Finish()
			if tc.mockSetup != nil {
				tc.mockSetup(mockAuthorUseCase)
			}
			resp, err := controller.GetAuthorInfo(context.Background(), tc.req)
			if tc.expectedError == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, authorID, resp.GetId())
				require.Equal(t, authorName, resp.GetName())
			} else {
				require.Error(t, err)
				require.Nil(t, resp)
				require.Equal(t, tc.expectedError, status.Code(err))
			}
		})
	}
}
