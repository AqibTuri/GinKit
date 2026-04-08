// Package domain holds persistence-oriented models (GORM). No HTTP imports here.
// Schema is defined in SQL migrations; structs must stay aligned with those tables.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User maps to the `users` table (see migrations). Keep domain models free of JSON tags when possible;
// use DTOs in handlers for API contracts.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Role         string    `gorm:"size:50;not null;default:user"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string {
	return "users"
}

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)
