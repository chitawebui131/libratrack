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

// @Summary      Create a new book
// @Description  Create a new book in the catalog
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        book  body  model.CreateBookRequest  true  "Book data"
// @Success      201   {object}  model.Book
// @Failure      422   {object}  map[string]interface{}
// @Router       /books [post]
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

// @Summary      List books
// @Description  Get list of books with optional filtering
// @Tags         books
// @Produce      json
// @Param        category  query  string  false  "Filter by category"
// @Param        page      query  int     false  "Page number"  default(1)
// @Param        limit     query  int     false  "Items per page"  default(10)
// @Success      200  {array}  model.Book
// @Router       /books [get]
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

// @Summary      Get a book by ID
// @Description  Get a single book by its ID
// @Tags         books
// @Produce      json
// @Param        id   path  int  true  "Book ID"
// @Success      200  {object}  model.Book
// @Failure      404  {object}  map[string]interface{}
// @Router       /books/{id} [get]
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

// @Summary      Update a book
// @Description  Update an existing book
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id    path  int  true  "Book ID"
// @Param        book  body  model.UpdateBookRequest  true  "Book data"
// @Success      200   {object}  model.Book
// @Failure      404   {object}  map[string]interface{}
// @Failure      422   {object}  map[string]interface{}
// @Router       /books/{id} [put]
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

// @Summary      Delete a book
// @Description  Delete a book by ID
// @Tags         books
// @Param        id  path  int  true  "Book ID"
// @Success      204  "No Content"
// @Failure      404  {object}  map[string]interface{}
// @Router       /books/{id} [delete]
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
