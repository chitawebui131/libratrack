package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"libratrack/internal/model"
	"libratrack/internal/repository"
)

type BookHandler struct {
	repo repository.BookRepository
}

func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

func (h *BookHandler) Create(c *gin.Context) {
	var req model.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{
				"code":    "validation_error",
				"message": "Validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	book := &model.Book{
		Title:         req.Title,
		ISBN:          req.ISBN,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	}

	if err := h.repo.Create(book); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "internal_error",
				"message": "Failed to create book",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, book)
}

func (h *BookHandler) List(c *gin.Context) {
	var category *string
	if c.Query("category") != "" {
		category = &c.Query("category")
	}

	page, _ := c.GetQueryInt("page")
	if page == 0 {
		page = 1
	}

	limit, _ := c.GetQueryInt("limit")
	if limit == 0 {
		limit = 10
	}

	books, err := h.repo.List(category, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "internal_error",
				"message": "Failed to list books",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, books)
}

func (h *BookHandler) Get(c *gin.Context) {
	id := c.Param("id")
	book, err := h.repo.FindByID(uint(id))
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "not_found",
					"message": "Book not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "internal_error",
					"message": "Failed to get book",
					"details": err.Error(),
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *BookHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{
				"code":    "validation_error",
				"message": "Validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	book := &model.Book{
		Title:         req.Title,
		ISBN:          req.ISBN,
		Author:        req.Author,
		Category:      req.Category,
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
	}

	if err := h.repo.Update(uint(id), book); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "not_found",
					"message": "Book not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "internal_error",
					"message": "Failed to update book",
					"details": err.Error(),
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *BookHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(uint(id)); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "not_found",
					"message": "Book not found",
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "internal_error",
					"message": "Failed to delete book",
					"details": err.Error(),
				},
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
