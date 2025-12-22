package validation

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MinUsernameLength = 3
	MaxUsernameLength = 50
)

var (
	// Email regex pattern (basic validation)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Username regex pattern (alphanumeric, underscore, hyphen)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return fmt.Errorf("email is required")
	}

	if len(email) > 255 {
		return fmt.Errorf("email is too long (max 255 characters)")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}

	if len(password) > MaxPasswordLength {
		return fmt.Errorf("password is too long (max %d characters)", MaxPasswordLength)
	}

	return nil
}

// ValidateUsername validates username format
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return fmt.Errorf("username is required")
	}

	if len(username) < MinUsernameLength {
		return fmt.Errorf("username must be at least %d characters", MinUsernameLength)
	}

	if len(username) > MaxUsernameLength {
		return fmt.Errorf("username is too long (max %d characters)", MaxUsernameLength)
	}

	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores and hyphens")
	}

	return nil
}

// ValidateTimezone validates timezone string
func ValidateTimezone(timezone string) error {
	if timezone == "" {
		return fmt.Errorf("timezone is required")
	}

	// Basic validation - just check it's not empty and reasonable length
	if len(timezone) > 50 {
		return fmt.Errorf("timezone is too long")
	}

	return nil
}
