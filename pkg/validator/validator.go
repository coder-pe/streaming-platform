// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	// emailRegex es una expresión regular para validar emails
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// ValidationError representa un error de validación
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewValidationError crea un nuevo error de validación
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// IsValidEmail valida si un email tiene un formato válido
func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// ValidateEmail valida un email y retorna un error si es inválido
func ValidateEmail(email string) error {
	if email == "" {
		return NewValidationError("email", "email is required")
	}
	if !IsValidEmail(email) {
		return NewValidationError("email", "invalid email format")
	}
	return nil
}

// ValidatePassword valida una contraseña con reglas de seguridad
func ValidatePassword(password string) error {
	if password == "" {
		return NewValidationError("password", "password is required")
	}

	if len(password) < 8 {
		return NewValidationError("password", "password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return NewValidationError("password", "password must not exceed 128 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return NewValidationError("password", "password must contain at least one uppercase letter")
	}
	if !hasLower {
		return NewValidationError("password", "password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return NewValidationError("password", "password must contain at least one number")
	}
	if !hasSpecial {
		return NewValidationError("password", "password must contain at least one special character")
	}

	return nil
}

// ValidatePasswordSimple valida una contraseña con reglas básicas (solo longitud)
func ValidatePasswordSimple(password string) error {
	if password == "" {
		return NewValidationError("password", "password is required")
	}
	if len(password) < 8 {
		return NewValidationError("password", "password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return NewValidationError("password", "password must not exceed 128 characters")
	}
	return nil
}

// ValidateRequired valida que un campo no esté vacío
func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(field, fmt.Sprintf("%s is required", field))
	}
	return nil
}

// ValidateMinLength valida la longitud mínima de un string
func ValidateMinLength(field, value string, minLength int) error {
	if len(strings.TrimSpace(value)) < minLength {
		return NewValidationError(field, fmt.Sprintf("%s must be at least %d characters", field, minLength))
	}
	return nil
}

// ValidateMaxLength valida la longitud máxima de un string
func ValidateMaxLength(field, value string, maxLength int) error {
	if len(value) > maxLength {
		return NewValidationError(field, fmt.Sprintf("%s must not exceed %d characters", field, maxLength))
	}
	return nil
}

// ValidateLength valida que un string esté entre min y max caracteres
func ValidateLength(field, value string, minLength, maxLength int) error {
	length := len(strings.TrimSpace(value))
	if length < minLength || length > maxLength {
		return NewValidationError(field, fmt.Sprintf("%s must be between %d and %d characters", field, minLength, maxLength))
	}
	return nil
}

// ValidateStringInSlice valida que un string esté en una lista de valores permitidos
func ValidateStringInSlice(field, value string, allowedValues []string) error {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return NewValidationError(field, fmt.Sprintf("%s must be one of: %v", field, allowedValues))
}

// NormalizeEmail normaliza un email (lowercase y trim)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeString normaliza un string (trim)
func NormalizeString(s string) string {
	return strings.TrimSpace(s)
}

// ValidateName valida un nombre (firstName, lastName)
func ValidateName(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return NewValidationError(field, fmt.Sprintf("%s is required", field))
	}
	if len(value) < 2 {
		return NewValidationError(field, fmt.Sprintf("%s must be at least 2 characters", field))
	}
	if len(value) > 50 {
		return NewValidationError(field, fmt.Sprintf("%s must not exceed 50 characters", field))
	}
	return nil
}

// ValidateVideoTitle valida el título de un video
func ValidateVideoTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return NewValidationError("title", "title is required")
	}
	if len(title) < 3 {
		return NewValidationError("title", "title must be at least 3 characters")
	}
	if len(title) > 200 {
		return NewValidationError("title", "title must not exceed 200 characters")
	}
	return nil
}

// ValidateVideoDescription valida la descripción de un video
func ValidateVideoDescription(description string) error {
	if len(description) > 5000 {
		return NewValidationError("description", "description must not exceed 5000 characters")
	}
	return nil
}

// ValidateCategory valida una categoría de video
func ValidateCategory(category string) error {
	allowedCategories := []string{
		"programming",
		"web-development",
		"mobile-development",
		"data-science",
		"devops",
		"design",
		"database",
		"security",
		"other",
	}
	return ValidateStringInSlice("category", category, allowedCategories)
}
