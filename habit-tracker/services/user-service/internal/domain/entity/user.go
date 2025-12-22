package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	Username      string    `json:"username" db:"username"`
	PasswordHash  string    `json:"-" db:"password_hash"`
	FirstName     *string   `json:"first_name,omitempty" db:"first_name"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	Timezone      string    `json:"timezone" db:"timezone"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// UserCreate represents data needed to create a new user
type UserCreate struct {
	Email     string  `json:"email" validate:"required,email"`
	Username  string  `json:"username" validate:"required,min=3,max=100"`
	Password  string  `json:"password" validate:"required,min=8"`
	FirstName *string `json:"first_name,omitempty"`
	Timezone  string  `json:"timezone" validate:"required"`
}

// UserUpdate represents data that can be updated
type UserUpdate struct {
	FirstName     *string `json:"first_name,omitempty"`
	Timezone      *string `json:"timezone,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
}

// UserResponse represents user data for API responses (without sensitive data)
type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	FirstName     *string   `json:"first_name,omitempty"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	Timezone      string    `json:"timezone"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		Username:      u.Username,
		FirstName:     u.FirstName,
		IsActive:      u.IsActive,
		EmailVerified: u.EmailVerified,
		Timezone:      u.Timezone,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}
