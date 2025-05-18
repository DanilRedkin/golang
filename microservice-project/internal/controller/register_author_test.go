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

func TestRegisterAuthor(t *testing.T) {
	t.Parallel()
	authorID := uuid.NewString()
	authorName := "New Author"

	testCases := []struct {
		name          string
		req           *library.RegisterAuthorRequest
		mockSetup     func(*lib.MockAuthorUseCase)
		expectedError codes.Code
	}{
		{
			name: "Success",
			req: &library.RegisterAuthorRequest{
				Name: authorName,
			},
			mockSetup: func(mockAuthorUseCase *lib.MockAuthorUseCase) {
				mockAuthorUseCase.EXPECT().
					RegisterAuthor(gomock.Any(), authorName).
					Return(entity.Author{ID: authorID, Name: authorName}, nil).
					Times(1)
			},
			expectedError: codes.OK,
		},
		{
			name: "AuthorAlreadyExists",
			req: &library.RegisterAuthorRequest{
				Name: "Existing Author",
			},
			mockSetup: func(mockAuthorUseCase *lib.MockAuthorUseCase) {
				mockAuthorUseCase.EXPECT().
					RegisterAuthor(gomock.Any(), "Existing Author").
					Return(entity.Author{}, entity.ErrAuthorAlreadyExists).
					Times(1)
			},
			expectedError: codes.AlreadyExists,
		},
		{
			name: "ValidationErrorEmptyName",
			req: &library.RegisterAuthorRequest{
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
			resp, err := controller.RegisterAuthor(context.Background(), tc.req)
			if tc.expectedError == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, authorID, resp.GetId())
			} else {
				require.Error(t, err)
				require.Nil(t, resp)
				require.Equal(t, tc.expectedError, status.Code(err))
			}
		})
	}
}
