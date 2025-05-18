package repository

import (
	"context"
	"sync"

	"github.com/project/library/internal/entity"
)

var _ AuthorRepository = (*impl)(nil)
var _ BooksRepository = (*impl)(nil)

type impl struct {
	authorsMx *sync.RWMutex
	authors   map[string]*entity.Author

	booksMx *sync.RWMutex
	books   map[string]*entity.Book
}

func New() *impl {
	return &impl{
		authorsMx: new(sync.RWMutex),
		authors:   make(map[string]*entity.Author),

		books:   map[string]*entity.Book{},
		booksMx: new(sync.RWMutex),
	}
}

func (i *impl) CreateBook(_ context.Context, book entity.Book) (entity.Book, error) {
	i.booksMx.Lock()
	defer i.booksMx.Unlock()

	if _, ok := i.books[book.ID]; ok {
		return entity.Book{}, entity.ErrBookAlreadyExists
	}

	if err := i.validateAuthorsExist(book.AuthorIDs); err != nil {
		return entity.Book{}, err
	}

	i.books[book.ID] = &book
	return book, nil
}

func (i *impl) GetBook(_ context.Context, bookID string) (entity.Book, error) {
	i.booksMx.RLock()
	defer i.booksMx.RUnlock()

	if requiredBook, ok := i.books[bookID]; !ok {
		return entity.Book{}, entity.ErrBookNotFound
	} else {
		return *requiredBook, nil
	}
}

func (i *impl) UpdateBook(_ context.Context, newBook entity.Book) error {
	i.booksMx.Lock()
	defer i.booksMx.Unlock()

	prevBook, ok := i.books[newBook.ID]
	if !ok {
		return entity.ErrBookNotFound
	}

	if err := i.validateAuthorsExist(newBook.AuthorIDs); err != nil {
		return err
	}

	prevBook.Name = newBook.Name
	prevBook.AuthorIDs = newBook.AuthorIDs
	return nil
}

func (i *impl) CreateAuthor(_ context.Context, author entity.Author) (entity.Author, error) {
	i.authorsMx.Lock()
	defer i.authorsMx.Unlock()

	if _, ok := i.authors[author.ID]; ok {
		return entity.Author{}, entity.ErrAuthorAlreadyExists
	}

	i.authors[author.ID] = &author
	return author, nil
}

func (i *impl) GetAuthor(_ context.Context, authorID string) (entity.Author, error) {
	i.authorsMx.RLock()
	defer i.authorsMx.RUnlock()

	if requiredAuthor, ok := i.authors[authorID]; !ok {
		return entity.Author{}, entity.ErrAuthorNotFound
	} else {
		return *requiredAuthor, nil
	}
}

func (i *impl) ChangeAuthor(_ context.Context, author entity.Author) error {
	i.authorsMx.Lock()
	defer i.authorsMx.Unlock()

	prevAuthor, ok := i.authors[author.ID]
	if !ok {
		return entity.ErrAuthorNotFound
	}
	prevAuthor.Name = author.Name
	return nil
}

func (i *impl) GetBooksByAuthor(_ context.Context, authorID string) ([]entity.Book, error) {
	i.booksMx.RLock()
	defer i.booksMx.RUnlock()

	i.authorsMx.RLock()
	_, authorExists := i.authors[authorID]
	i.authorsMx.RUnlock()

	if !authorExists {
		return nil, entity.ErrAuthorNotFound
	}

	var books []entity.Book
	for _, book := range i.books {
		for _, id := range book.AuthorIDs {
			if id == authorID {
				books = append(books, *book)
				break
			}
		}
	}

	if len(books) == 0 {
		return nil, entity.ErrBookNotFound
	}

	return books, nil
}

func (i *impl) validateAuthorsExist(authorIDs []string) error {
	i.authorsMx.RLock()
	defer i.authorsMx.RUnlock()

	for _, authorID := range authorIDs {
		if _, exists := i.authors[authorID]; !exists {
			return entity.ErrAuthorNotFound
		}
	}
	return nil
}
