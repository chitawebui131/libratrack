package repository

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateAndFindByID(t *testing.T) {
	repo := NewInMemoryBookRepository()

	book := &model.Book{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(book)
	assert.NoError(t, err)

	fetchedBook, err := repo.FindByID(book.ID)
	assert.NoError(t, err)
	assert.Equal(t, book.ID, fetchedBook.ID)
	assert.Equal(t, book.Title, fetchedBook.Title)
	assert.Equal(t, book.ISBN, fetchedBook.ISBN)
	assert.Equal(t, book.Author, fetchedBook.Author)
	assert.Equal(t, book.Category, fetchedBook.Category)
	assert.Equal(t, book.PublishedYear, fetchedBook.PublishedYear)
	assert.Equal(t, book.Description, fetchedBook.Description)
}

func TestUpdateExistingBook(t *testing.T) {
	repo := NewInMemoryBookRepository()

	book := &model.Book{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(book)
	assert.NoError(t, err)

	updatedBook := &model.Book{
		Title:         "Updated Book",
		ISBN:          "0987654321",
		Author:        "Updated Author",
		Category:      "Updated Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err = repo.Update(book.ID, updatedBook)
	assert.NoError(t, err)

	fetchedBook, err := repo.FindByID(book.ID)
	assert.NoError(t, err)
	assert.Equal(t, updatedBook.Title, fetchedBook.Title)
	assert.Equal(t, updatedBook.ISBN, fetchedBook.ISBN)
	assert.Equal(t, updatedBook.Author, fetchedBook.Author)
	assert.Equal(t, updatedBook.Category, fetchedBook.Category)
	assert.Equal(t, updatedBook.PublishedYear, fetchedBook.PublishedYear)
	assert.Equal(t, updatedBook.Description, fetchedBook.Description)
}

func TestDeleteExistingBook(t *testing.T) {
	repo := NewInMemoryBookRepository()

	book := &model.Book{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(book)
	assert.NoError(t, err)

	err = repo.Delete(book.ID)
	assert.NoError(t, err)

	_, err = repo.FindByID(book.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestListWithCategoryFilter(t *testing.T) {
	repo := NewInMemoryBookRepository()

	book1 := &model.Book{
		Title:         "Test Book 1",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	book2 := &model.Book{
		Title:         "Test Book 2",
		ISBN:          "0987654321",
		Author:        "Test Author",
		Category:      "Another Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(book1)
	assert.NoError(t, err)

	err = repo.Create(book2)
	assert.NoError(t, err)

	books, err := repo.List(&book1.Category, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, books, 1)
	assert.Equal(t, book1.ID, books[0].ID)
}

func TestNotFoundCases(t *testing.T) {
	repo := NewInMemoryBookRepository()

	book := &model.Book{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(book)
	assert.NoError(t, err)

	_, err = repo.FindByID(999)
	assert.ErrorIs(t, err, ErrNotFound)

	err = repo.Update(999, book)
	assert.ErrorIs(t, err, ErrNotFound)

	err = repo.Delete(999)
	assert.ErrorIs(t, err, ErrNotFound)
}
