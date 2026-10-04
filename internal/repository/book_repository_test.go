package repository

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBook(t *testing.T) {
	repo := NewInMemoryBookRepository()

	req := model.CreateBookRequest{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(&model.Book{
		Title:         req.Title,
		ISBN:          req.ISBN,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	})

	require.NoError(t, err)
}

func TestFindByID(t *testing.T) {
	repo := NewInMemoryBookRepository()

	req := model.CreateBookRequest{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(&model.Book{
		Title:         req.Title,
		ISBN:          req.ISBN,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	})

	require.NoError(t, err)

	book, err := repo.FindByID(1)
	require.NoError(t, err)
	assert.Equal(t, "Test Book", book.Title)
}

func TestUpdateBook(t *testing.T) {
	repo := NewInMemoryBookRepository()

	req := model.CreateBookRequest{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(&model.Book{
		Title:         req.Title,
		ISBN:          req.ISBN,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	})

	require.NoError(t, err)

	updateReq := model.UpdateBookRequest{
		Title:         "Updated Book",
		ISBN:          "0987654321",
		Author:        "Updated Author",
		Category:      "Updated Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err = repo.Update(1, &model.Book{
		Title:         updateReq.Title,
		ISBN:          updateReq.ISBN,
		Author:        updateReq.Author,
		Category:      updateReq.Category,
		PublishedYear: updateReq.PublishedYear,
		Description:   updateReq.Description,
	})

	require.NoError(t, err)

	book, err := repo.FindByID(1)
	require.NoError(t, err)
	assert.Equal(t, "Updated Book", book.Title)
}

func TestDeleteBook(t *testing.T) {
	repo := NewInMemoryBookRepository()

	req := model.CreateBookRequest{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(&model.Book{
		Title:         req.Title,
		ISBN:          req.ISBN,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	})

	require.NoError(t, err)

	err = repo.Delete(1)
	require.NoError(t, err)

	_, err = repo.FindByID(1)
	assert.Equal(t, ErrNotFound, err)
}

func TestListBooks(t *testing.T) {
	repo := NewInMemoryBookRepository()

	req := model.CreateBookRequest{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	err := repo.Create(&model.Book{
		Title:         req.Title,
		ISBN:          req.ISBN,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	})

	require.NoError(t, err)

	books, err := repo.List(&"Test Category", 1, 10)
	require.NoError(t, err)
	assert.Len(t, books, 1)
}

func TestFindByIDNotFound(t *testing.T) {
	repo := NewInMemoryBookRepository()

	_, err := repo.FindByID(1)
	assert.Equal(t, ErrNotFound, err)
}

func TestUpdateBookNotFound(t *testing.T) {
	repo := NewInMemoryBookRepository()

	err := repo.Update(1, &model.Book{})
	assert.Equal(t, ErrNotFound, err)
}

func TestDeleteBookNotFound(t *testing.T) {
	repo := NewInMemoryBookRepository()

	err := repo.Delete(1)
	assert.Equal(t, ErrNotFound, err)
}
