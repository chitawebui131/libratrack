package repository

import (
	"errors"
	"sync"
	"time"

	"libratrack/internal/model"
)

var ErrNotFound = errors.New("not found")

type BookRepository interface {
	Create(book *model.Book) error
	FindByID(id uint) (*model.Book, error)
	List(category *string, page, limit int) ([]*model.Book, error)
	Update(id uint, book *model.Book) error
	Delete(id uint) error
}

type inMemoryBookRepository struct {
	books  map[uint]*model.Book
	nextID uint
	mu     sync.RWMutex
}

func NewInMemoryBookRepository() BookRepository {
	return &inMemoryBookRepository{
		books:  make(map[uint]*model.Book),
		nextID: 1,
	}
}

func (r *inMemoryBookRepository) Create(book *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	book.ID = r.nextID
	r.books[book.ID] = book
	r.nextID++
	return nil
}

func (r *inMemoryBookRepository) FindByID(id uint) (*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book, exists := r.books[id]
	if !exists {
		return nil, ErrNotFound
	}
	return book, nil
}

func (r *inMemoryBookRepository) List(category *string, page, limit int) ([]*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var books []*model.Book
	for _, book := range r.books {
		if category == nil || book.Category == *category {
			books = append(books, book)
		}
	}

	start := (page - 1) * limit
	if start >= len(books) {
		return nil, nil
	}
	if start+limit > len(books) {
		return books[start:], nil
	}
	return books[start : start+limit], nil
}

func (r *inMemoryBookRepository) Update(id uint, book *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existingBook, exists := r.books[id]
	if !exists {
		return ErrNotFound
	}

	existingBook.Title = book.Title
	existingBook.ISBN = book.ISBN
	existingBook.Author = book.Author
	existingBook.Category = book.Category
	existingBook.PublishedYear = book.PublishedYear
	existingBook.Description = book.Description
	existingBook.UpdatedAt = time.Now()
	return nil
}

func (r *inMemoryBookRepository) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.books[id]
	if !exists {
		return ErrNotFound
	}
	delete(r.books, id)
	return nil
}
