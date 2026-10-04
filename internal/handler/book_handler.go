package handler

import (
	"net/http"
	"strconv"

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

func (h *BookHandler) CreateBook(c *gin.Context) {
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
				"code":    "internal_server_error",
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, book)
}

func (h *BookHandler) ListBooks(c *gin.Context) {
	category := c.Query("category")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	books, err := h.repo.List(&category, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "internal_server_error",
				"message": "Internal server error",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, books)
}

func (h *BookHandler) GetBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "invalid_id",
				"message": "Invalid ID",
				"details": err.Error(),
			},
		})
		return
	}

	book, err := h.repo.FindByID(uint(id))
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "not_found",
					"message": "Book not found",
					"details": err.Error(),
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "internal_server_error",
					"message": "Internal server error",
					"details": err.Error(),
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *BookHandler) UpdateBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "invalid_id",
				"message": "Invalid ID",
				"details": err.Error(),
			},
		})
		return
	}

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
					"details": err.Error(),
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "internal_server_error",
					"message": "Internal server error",
					"details": err.Error(),
				},
			})
		}
		return
	}

	c.JSON(http.StatusOK, book)
}

func (h *BookHandler) DeleteBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "invalid_id",
				"message": "Invalid ID",
				"details": err.Error(),
			},
		})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "not_found",
					"message": "Book not found",
					"details": err.Error(),
				},
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "internal_server_error",
					"message": "Internal server error",
					"details": err.Error(),
				},
			})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
