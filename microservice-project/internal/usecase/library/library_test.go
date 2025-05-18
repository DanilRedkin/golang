package library_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/project/library/internal/entity"
	"github.com/project/library/internal/usecase/library"
	"github.com/project/library/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (
	*gomock.Controller,
	*mocks.MockAuthorRepository,
	*mocks.MockBooksRepository,
	*mocks.MockOutboxRepository,
	*library.Impl,
) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
	mockBookRepo := mocks.NewMockBooksRepository(ctrl)
	mockOutboxRepo := mocks.NewMockOutboxRepository(ctrl)
	mockTransactor := mocks.NewMockTransactor(ctrl)

	mockTransactor.
		EXPECT().
		WithTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).
		AnyTimes()

	useCase := library.New(
		zap.NewNop(),
		mockAuthorRepo,
		mockBookRepo,
		mockOutboxRepo,
		mockTransactor,
	)
	return ctrl, mockAuthorRepo, mockBookRepo, mockOutboxRepo, useCase
}

func TestAuthorUseCaseTests(t *testing.T) {
	t.Parallel()

	t.Run("RegisterAuthor", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name          string
			authorName    string
			setupMocks    func(*mocks.MockAuthorRepository, *mocks.MockOutboxRepository)
			expectedError error
		}{
			{
				name:       "Success",
				authorName: "Alexander Pushkin",
				setupMocks: func(authorRepo *mocks.MockAuthorRepository, outboxRepo *mocks.MockOutboxRepository) {
					authorRepo.
						EXPECT().
						CreateAuthor(gomock.Any(), gomock.AssignableToTypeOf(entity.Author{})).
						DoAndReturn(func(_ context.Context, a entity.Author) (entity.Author, error) {
							a.ID = uuid.NewString()
							return a, nil
						})
					outboxRepo.
						EXPECT().
						SendMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				expectedError: nil,
			},
			{
				name:       "ErrorAuthorAlreadyExists",
				authorName: "Alexander Pushkin",
				setupMocks: func(authorRepo *mocks.MockAuthorRepository, _ *mocks.MockOutboxRepository) {
					authorRepo.
						EXPECT().
						CreateAuthor(gomock.Any(), gomock.Any()).
						Return(entity.Author{}, entity.ErrAuthorAlreadyExists)
				},
				expectedError: entity.ErrAuthorAlreadyExists,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ctrl, authorRepo, _, outboxRepo, useCase := setupTest(t)
				defer ctrl.Finish()

				tc.setupMocks(authorRepo, outboxRepo)
				ctx := context.Background()
				author, err := useCase.RegisterAuthor(ctx, tc.authorName)

				if tc.expectedError == nil {
					require.NoError(t, err)
					require.NotEmpty(t, author.ID)
					require.Equal(t, tc.authorName, author.Name)
				} else {
					require.Error(t, err)
					require.ErrorIs(t, err, tc.expectedError)
				}
			})
		}
	})

	t.Run("GetAuthor", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name          string
			authorID      string
			setupMocks    func(*mocks.MockAuthorRepository, string)
			expectedError error
		}{
			{
				name:     "Success",
				authorID: uuid.NewString(),
				setupMocks: func(authorRepo *mocks.MockAuthorRepository, id string) {
					authorRepo.
						EXPECT().
						GetAuthor(gomock.Any(), gomock.Eq(id)).
						Return(entity.Author{ID: id, Name: "Alexander Pushkin"}, nil)
				},
				expectedError: nil,
			},
			{
				name:     "ErrorAuthorNotFound",
				authorID: uuid.NewString(),
				setupMocks: func(authorRepo *mocks.MockAuthorRepository, id string) {
					authorRepo.
						EXPECT().
						GetAuthor(gomock.Any(), gomock.Eq(id)).
						Return(entity.Author{}, entity.ErrAuthorNotFound)
				},
				expectedError: entity.ErrAuthorNotFound,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ctrl, authorRepo, _, _, useCase := setupTest(t)
				defer ctrl.Finish()

				tc.setupMocks(authorRepo, tc.authorID)
				ctx := context.Background()
				author, err := useCase.GetAuthor(ctx, tc.authorID)

				if tc.expectedError == nil {
					require.NoError(t, err)
					require.Equal(t, "Alexander Pushkin", author.Name)
				} else {
					require.Error(t, err)
					require.ErrorIs(t, err, tc.expectedError)
				}
			})
		}
	})

	t.Run("ChangeAuthor", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name          string
			authorID      string
			newName       string
			setupMocks    func(*mocks.MockAuthorRepository, string, string)
			expectedError error
		}{
			{
				name:     "Success",
				authorID: uuid.NewString(),
				newName:  "Alexander Pushkin",
				setupMocks: func(authorRepo *mocks.MockAuthorRepository, id, name string) {
					authorRepo.
						EXPECT().
						ChangeAuthor(gomock.Any(), gomock.AssignableToTypeOf(entity.Author{})).
						DoAndReturn(func(_ context.Context, a entity.Author) error {
							if a.ID != id || a.Name != name {
								t.Fatalf("unexpected author passed to ChangeAuthor: %+v", a)
							}
							return nil
						})
				},
				expectedError: nil,
			},
			{
				name:     "ErrorAuthorNotFound",
				authorID: uuid.NewString(),
				newName:  "Alexander Pushkin",
				setupMocks: func(authorRepo *mocks.MockAuthorRepository, id, name string) {
					authorRepo.
						EXPECT().
						ChangeAuthor(gomock.Any(), gomock.AssignableToTypeOf(entity.Author{})).
						Return(entity.ErrAuthorNotFound)
				},
				expectedError: entity.ErrAuthorNotFound,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ctrl, authorRepo, _, _, useCase := setupTest(t)
				defer ctrl.Finish()

				tc.setupMocks(authorRepo, tc.authorID, tc.newName)
				ctx := context.Background()
				err := useCase.ChangeAuthor(ctx, tc.authorID, tc.newName)

				if tc.expectedError == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					require.ErrorIs(t, err, tc.expectedError)
				}
			})
		}
	})
}

func TestBooksUseCaseTests(t *testing.T) {
	t.Parallel()

	t.Run("RegisterBook", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name          string
			bookName      string
			authorIDs     []string
			setupMocks    func(*mocks.MockBooksRepository, *mocks.MockOutboxRepository)
			expectedError error
		}{
			{
				name:      "Success",
				bookName:  "Evgeny Onegin",
				authorIDs: []string{uuid.NewString()},
				setupMocks: func(bookRepo *mocks.MockBooksRepository, outboxRepo *mocks.MockOutboxRepository) {
					bookRepo.
						EXPECT().
						CreateBook(gomock.Any(), gomock.AssignableToTypeOf(entity.Book{})).
						DoAndReturn(func(_ context.Context, b entity.Book) (entity.Book, error) {
							b.ID = uuid.NewString()
							return b, nil
						})
					outboxRepo.
						EXPECT().
						SendMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
				},
				expectedError: nil,
			},
			{
				name:      "ErrorBookAlreadyExists",
				bookName:  "Evgeny Onegin",
				authorIDs: []string{uuid.NewString()},
				setupMocks: func(bookRepo *mocks.MockBooksRepository, _ *mocks.MockOutboxRepository) {
					bookRepo.
						EXPECT().
						CreateBook(gomock.Any(), gomock.Any()).
						Return(entity.Book{}, entity.ErrBookAlreadyExists)
				},
				expectedError: entity.ErrBookAlreadyExists,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ctrl, _, bookRepo, outboxRepo, useCase := setupTest(t)
				defer ctrl.Finish()

				tc.setupMocks(bookRepo, outboxRepo)
				ctx := context.Background()
				newBook := entity.Book{Name: tc.bookName, AuthorIDs: tc.authorIDs}
				book, err := useCase.RegisterBook(ctx, newBook)

				if tc.expectedError == nil {
					require.NoError(t, err)
					require.NotEmpty(t, book.ID)
					require.Equal(t, tc.bookName, book.Name)
					require.Equal(t, tc.authorIDs, book.AuthorIDs)
				} else {
					require.Error(t, err)
					require.ErrorIs(t, err, tc.expectedError)
				}
			})
		}
	})

	t.Run("GetBook", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name          string
			bookID        string
			setupMocks    func(*mocks.MockBooksRepository, string)
			expectedError error
		}{
			{
				name:   "Success",
				bookID: uuid.NewString(),
				setupMocks: func(bookRepo *mocks.MockBooksRepository, id string) {
					bookRepo.
						EXPECT().
						GetBook(gomock.Any(), gomock.Eq(id)).
						Return(entity.Book{ID: id, Name: "Evgeny Onegin"}, nil)
				},
				expectedError: nil,
			},
			{
				name:   "ErrorBookNotFound",
				bookID: uuid.NewString(),
				setupMocks: func(bookRepo *mocks.MockBooksRepository, id string) {
					bookRepo.
						EXPECT().
						GetBook(gomock.Any(), gomock.Eq(id)).
						Return(entity.Book{}, entity.ErrBookNotFound)
				},
				expectedError: entity.ErrBookNotFound,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ctrl, _, bookRepo, _, useCase := setupTest(t)
				defer ctrl.Finish()

				tc.setupMocks(bookRepo, tc.bookID)
				ctx := context.Background()
				book, err := useCase.GetBook(ctx, tc.bookID)

				if tc.expectedError == nil {
					require.NoError(t, err)
					require.Equal(t, "Evgeny Onegin", book.Name)
				} else {
					require.Error(t, err)
					require.ErrorIs(t, err, tc.expectedError)
				}
			})
		}
	})

	t.Run("UpdateBook", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name          string
			bookID        string
			bookName      string
			authorIDs     []string
			setupMocks    func(*mocks.MockBooksRepository)
			expectedError error
		}{
			{
				name:      "Success",
				bookID:    uuid.NewString(),
				bookName:  "Evgeny Onegin",
				authorIDs: []string{uuid.NewString()},
				setupMocks: func(bookRepo *mocks.MockBooksRepository) {
					bookRepo.
						EXPECT().
						UpdateBook(gomock.Any(), gomock.AssignableToTypeOf(entity.Book{})).
						Return(nil)
				},
				expectedError: nil,
			},
			{
				name:      "ErrorBookNotFound",
				bookID:    uuid.NewString(),
				bookName:  "Evgeny Onegin",
				authorIDs: []string{uuid.NewString()},
				setupMocks: func(bookRepo *mocks.MockBooksRepository) {
					bookRepo.
						EXPECT().
						UpdateBook(gomock.Any(), gomock.AssignableToTypeOf(entity.Book{})).
						Return(entity.ErrBookNotFound)
				},
				expectedError: entity.ErrBookNotFound,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ctrl, _, bookRepo, _, useCase := setupTest(t)
				defer ctrl.Finish()

				tc.setupMocks(bookRepo)
				ctx := context.Background()
				err := useCase.UpdateBook(ctx, tc.bookID, tc.bookName, tc.authorIDs)

				if tc.expectedError == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					require.ErrorIs(t, err, tc.expectedError)
				}
			})
		}
	})

	t.Run("GetAuthorBooks", func(t *testing.T) {
		t.Parallel()

		authorID := uuid.NewString()
		book1 := entity.Book{ID: uuid.NewString(), Name: "Evgeny Onegin", AuthorIDs: []string{authorID}}
		book2 := entity.Book{ID: uuid.NewString(), Name: "Evgeny Onegin 2", AuthorIDs: []string{authorID}}

		testCases := []struct {
			name          string
			authorID      string
			setupMocks    func(*mocks.MockBooksRepository, string)
			expectedBooks []entity.Book
			expectedError error
		}{
			{
				name:     "Success",
				authorID: authorID,
				setupMocks: func(bookRepo *mocks.MockBooksRepository, id string) {
					bookRepo.
						EXPECT().
						GetBooksByAuthor(gomock.Any(), gomock.Eq(id)).
						Return([]entity.Book{book1, book2}, nil)
				},
				expectedBooks: []entity.Book{book1, book2},
				expectedError: nil,
			},
			{
				name:     "ErrorAuthorNotFound",
				authorID: authorID,
				setupMocks: func(bookRepo *mocks.MockBooksRepository, id string) {
					bookRepo.
						EXPECT().
						GetBooksByAuthor(gomock.Any(), gomock.Eq(id)).
						Return(nil, entity.ErrAuthorNotFound)
				},
				expectedBooks: nil,
				expectedError: entity.ErrAuthorNotFound,
			},
			{
				name:     "ErrorBookNotFound",
				authorID: authorID,
				setupMocks: func(bookRepo *mocks.MockBooksRepository, id string) {
					bookRepo.
						EXPECT().
						GetBooksByAuthor(gomock.Any(), gomock.Eq(id)).
						Return(nil, entity.ErrBookNotFound)
				},
				expectedBooks: nil,
				expectedError: entity.ErrBookNotFound,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ctrl, _, bookRepo, _, useCase := setupTest(t)
				defer ctrl.Finish()

				tc.setupMocks(bookRepo, tc.authorID)
				ctx := context.Background()
				books, err := useCase.GetAuthorBooks(ctx, tc.authorID)
				if tc.expectedError == nil {
					require.NoError(t, err)
					require.Equal(t, tc.expectedBooks, books)
				} else {
					require.Error(t, err)
					require.ErrorIs(t, err, tc.expectedError)
				}
			})
		}
	})
}
