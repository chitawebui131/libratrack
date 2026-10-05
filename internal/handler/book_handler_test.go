package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBook(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	h := NewBookHandler(repo)

	req := model.CreateBookRequest{
		Title:         "Test Book",
		ISBN:          "1234567890",
		Author:        "Test Author",
		Category:      "Test Category",
		PublishedYear: new(int),
		Description:   new(string),
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/v1/books", gin.H{
		"title":          req.Title,
		"isbn":           req.ISBN,
		"author":         req.Author,
		"category":       req.Category,
		"published_year": req.PublishedYear,
		"description":    req.Description,
	})

	h.CreateBook(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetBook(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	h := NewBookHandler(repo)

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

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/books/1", nil)

	h.GetBook(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateBook(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	h := NewBookHandler(repo)

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

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("PUT", "/api/v1/books/1", gin.H{
		"title":          updateReq.Title,
		"isbn":           updateReq.ISBN,
		"author":         updateReq.Author,
		"category":       updateReq.Category,
		"published_year": updateReq.PublishedYear,
		"description":    updateReq.Description,
	})

	h.UpdateBook(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteBook(t *testing.T) {
	repo := repository.NewInMemoryBookRepository()
	h := NewBookHandler(repo)

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

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", "/api/v1/books/1", nil)

	h.DeleteBook(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
