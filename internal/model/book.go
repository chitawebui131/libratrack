package model

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	// binding:"required" accepts whitespace-only strings, so we add the "nonblank" tag.
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("nonblank", func(fl validator.FieldLevel) bool {
			return strings.TrimSpace(fl.Field().String()) != ""
		})
	}
}

// Author and Category are temporary plain-text fields and will become relational in theme 8.
type Book struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Title         string    `json:"title" gorm:"not null"`
	ISBN          string    `json:"isbn" gorm:"not null"`
	Author        string    `json:"author" gorm:"not null"`
	Category      string    `json:"category" gorm:"not null"`
	PublishedYear *int      `json:"published_year" gorm:"default:null"`
	Description   *string   `json:"description" gorm:"default:null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CreateBookRequest struct {
	Title         string  `json:"title" binding:"required,nonblank"`
	ISBN          string  `json:"isbn" binding:"required,nonblank"`
	Author        string  `json:"author" binding:"required,nonblank"`
	Category      string  `json:"category" binding:"required,nonblank"`
	PublishedYear *int    `json:"published_year,omitempty"`
	Description   *string `json:"description,omitempty"`
}

type UpdateBookRequest struct {
	Title         string  `json:"title" binding:"required,nonblank"`
	ISBN          string  `json:"isbn" binding:"required,nonblank"`
	Author        string  `json:"author" binding:"required,nonblank"`
	Category      string  `json:"category" binding:"required,nonblank"`
	PublishedYear *int    `json:"published_year,omitempty"`
	Description   *string `json:"description,omitempty"`
}