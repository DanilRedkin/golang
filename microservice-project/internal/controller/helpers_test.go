package controller

import (
	"testing"

	"github.com/golang/mock/gomock"
	lib "github.com/project/library/mocks"
	"go.uber.org/zap"
)

var logger = zap.NewNop()

func setupTest(t *testing.T) (*gomock.Controller, *Implementation, *lib.MockBooksUseCase, *lib.MockAuthorUseCase) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockBooksUseCase := lib.NewMockBooksUseCase(ctrl)
	mockAuthorUseCase := lib.NewMockAuthorUseCase(ctrl)
	controller := New(logger, mockBooksUseCase, mockAuthorUseCase)
	return ctrl, controller, mockBooksUseCase, mockAuthorUseCase
}
