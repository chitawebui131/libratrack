package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"libratrack/internal/model"
	"libratrack/internal/repository"
)

const maxPageLimit = 100

type BookHandler struct {
	repo repository.BookRepository
}

func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

func respondError(c *gin.Context, status int, code, message string, details ...string) {
	body := gin.H{"code": code, "message": message}
	if len(details) > 0 {
		body["details"] = details[0]
	}
	c.JSON(status, gin.H{"error": body})
}

// respondRepoError returns 404 for ErrNotFound; other errors are logged and hidden from the client.
func respondRepoError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		respondError(c, http.StatusNotFound, "not_found", "Book not found")
		return
	}
	log.Printf("repository error: %v", err)
	respondError(c, http.StatusInternalServerError, "internal_server_error", "Internal server error")
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "Invalid ID", err.Error())
		return 0, false
	}
	return uint(id), true
}

func parsePositiveQueryInt(c *gin.Context, name, def string) (int, bool) {
	v, err := strconv.Atoi(c.DefaultQuery(name, def))
	if err != nil || v < 1 {
		respondError(c, http.StatusBadRequest, "invalid_pagination", name+" must be a positive integer")
		return 0, false
	}
	return v, true
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
		respondError(c, http.StatusUnprocessableEntity, "validation_error", "Validation failed", err.Error())
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
		respondRepoError(c, err)
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
// @Param        limit     query  int     false  "Items per page (max 100)"  default(10)
// @Success      200  {array}  model.Book
// @Failure      400  {object}  map[string]interface{}
// @Router       /books [get]
func (h *BookHandler) ListBooks(c *gin.Context) {
	page, ok := parsePositiveQueryInt(c, "page", "1")
	if !ok {
		return
	}
	limit, ok := parsePositiveQueryInt(c, "limit", "10")
	if !ok {
		return
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}

	var category *string
	if v := c.Query("category"); v != "" {
		category = &v
	}

	books, err := h.repo.List(category, page, limit)
	if err != nil {
		respondRepoError(c, err)
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
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /books/{id} [get]
func (h *BookHandler) GetBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	book, err := h.repo.FindByID(id)
	if err != nil {
		respondRepoError(c, err)
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
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Failure      422   {object}  map[string]interface{}
// @Router       /books/{id} [put]
func (h *BookHandler) UpdateBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req model.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusUnprocessableEntity, "validation_error", "Validation failed", err.Error())
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

	if err := h.repo.Update(id, book); err != nil {
		respondRepoError(c, err)
		return
	}

	// Return the stored book (with ID and timestamps), not the object built from the request.
	updated, err := h.repo.FindByID(id)
	if err != nil {
		respondRepoError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

// @Summary      Delete a book
// @Description  Delete a book by ID
// @Tags         books
// @Param        id  path  int  true  "Book ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /books/{id} [delete]
func (h *BookHandler) DeleteBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.repo.Delete(id); err != nil {
		respondRepoError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}