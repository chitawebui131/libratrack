package repository

import (
	"errors"
	"sort"
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

// cloneBook makes a deep copy so the store never shares memory with callers.
func cloneBook(b *model.Book) *model.Book {
	c := *b
	if b.PublishedYear != nil {
		year := *b.PublishedYear
		c.PublishedYear = &year
	}
	if b.Description != nil {
		desc := *b.Description
		c.Description = &desc
	}
	return &c
}

func (r *inMemoryBookRepository) Create(book *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	book.ID = r.nextID
	book.CreatedAt = now
	book.UpdatedAt = now

	r.books[book.ID] = cloneBook(book)
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
	return cloneBook(book), nil
}

// List returns books sorted by ID. A nil or empty category means "no filter".
// page < 1 is treated as 1; limit <= 0 returns an empty result.
func (r *inMemoryBookRepository) List(category *string, page, limit int) ([]*model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filtered := make([]*model.Book, 0, len(r.books))
	for _, book := range r.books {
		if category == nil || *category == "" || book.Category == *category {
			filtered = append(filtered, book)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })

	result := make([]*model.Book, 0)
	if limit <= 0 {
		return result, nil
	}
	if page < 1 {
		page = 1
	}
	// Check before multiplying so (page-1)*limit cannot overflow.
	if page-1 > len(filtered)/limit {
		return result, nil
	}
	start := (page - 1) * limit
	if start >= len(filtered) {
		return result, nil
	}
	end := len(filtered)
	if limit < len(filtered)-start {
		end = start + limit
	}

	for _, book := range filtered[start:end] {
		result = append(result, cloneBook(book))
	}
	return result, nil
}

func (r *inMemoryBookRepository) Update(id uint, book *model.Book) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.books[id]
	if !exists {
		return ErrNotFound
	}

	incoming := cloneBook(book)
	existing.Title = incoming.Title
	existing.ISBN = incoming.ISBN
	existing.Author = incoming.Author
	existing.Category = incoming.Category
	existing.PublishedYear = incoming.PublishedYear
	existing.Description = incoming.Description
	existing.UpdatedAt = time.Now()
	return nil
}

func (r *inMemoryBookRepository) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.books[id]; !exists {
		return ErrNotFound
	}
	delete(r.books, id)
	return nil
}