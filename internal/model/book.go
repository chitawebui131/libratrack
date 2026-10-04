package model

import (
	"time"
)

// Author and Category are temporary plain-text fields and will become relational in theme 8.
type Book struct {
	ID            uint      `gorm:"primaryKey"`
	Title         string    `gorm:"not null"`
	ISBN          string    `gorm:"not null"`
	Author        string    `gorm:"not null"`
	Category      string    `gorm:"not null"`
	PublishedYear *int      `gorm:"default:null"`
	Description   *string   `gorm:"default:null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

type CreateBookRequest struct {
	Title         string  `json:"title" binding:"required"`
	ISBN          string  `json:"isbn" binding:"required"`
	Author        string  `json:"author" binding:"required"`
	Category      string  `json:"category" binding:"required"`
	PublishedYear *int    `json:"published_year,omitempty"`
	Description   *string `json:"description,omitempty"`
}

type UpdateBookRequest struct {
	Title         string  `json:"title,omitempty" binding:"required"`
	ISBN          string  `json:"isbn,omitempty" binding:"required"`
	Author        string  `json:"author,omitempty" binding:"required"`
	Category      string  `json:"category,omitempty" binding:"required"`
	PublishedYear *int    `json:"published_year,omitempty"`
	Description   *string `json:"description,omitempty"`
}
