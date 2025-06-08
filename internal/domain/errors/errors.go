package errors

import (
	"errors"
	"fmt"
)

// Domain error types
var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrWeakPassword       = errors.New("password is too weak")
	ErrPasswordMismatch   = errors.New("passwords do not match")

	// Video errors
	ErrVideoNotFound      = errors.New("video not found")
	ErrVideoNotReady      = errors.New("video is not ready for streaming")
	ErrVideoProcessing    = errors.New("video is still processing")
	ErrVideoUploadFailed  = errors.New("video upload failed")
	ErrInvalidVideoFormat = errors.New("invalid video format")
	ErrVideoTooLarge      = errors.New("video file too large")
	ErrVideoAccess        = errors.New("access denied to video")

	// Auth errors
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token has expired")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrSessionExpired      = errors.New("session has expired")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrForbidden           = errors.New("forbidden access")

	// Validation errors
	ErrInvalidInput    = errors.New("invalid input")
	ErrMissingField    = errors.New("required field is missing")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUUID     = errors.New("invalid UUID format")
	ErrInvalidFileType = errors.New("invalid file type")

	// Database errors
	ErrDatabaseConnection  = errors.New("database connection error")
	ErrDatabaseQuery       = errors.New("database query error")
	ErrDatabaseTransaction = errors.New("database transaction error")
	ErrRecordNotFound      = errors.New("record not found")
	ErrDuplicateRecord     = errors.New("duplicate record")

	// Cache errors
	ErrCacheConnection  = errors.New("cache connection error")
	ErrCacheKeyNotFound = errors.New("cache key not found")
	ErrCacheSetFailed   = errors.New("failed to set cache")

	// External service errors
	ErrExternalService    = errors.New("external service error")
	ErrEmailService       = errors.New("email service error")
	ErrStorageService     = errors.New("storage service error")
	ErrTranscodingService = errors.New("transcoding service error")

	// Rate limiting errors
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrTooManyRequests   = errors.New("too many requests")

	// Business logic errors
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrResourceNotAvailable    = errors.New("resource not available")
	ErrConcurrentModification  = errors.New("concurrent modification detected")
	ErrQuotaExceeded           = errors.New("quota exceeded")
)

// DomainError represents a domain-specific error with additional context
type DomainError struct {
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Context map[string]interface{} `json:"context,omitempty"`
	Err     error                  `json:"-"`
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap implements the unwrap interface for error chains
func (e *DomainError) Unwrap() error {
	return e.Err
}

// Is implements the Is interface for error comparison
func (e *DomainError) Is(target error) bool {
	if t, ok := target.(*DomainError); ok {
		return e.Type == t.Type
	}
	return errors.Is(e.Err, target)
}

// NewDomainError creates a new domain error
func NewDomainError(errorType, message string, code int) *DomainError {
	return &DomainError{
		Type:    errorType,
		Message: message,
		Code:    code,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *DomainError) WithContext(key string, value interface{}) *DomainError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithError wraps an underlying error
func (e *DomainError) WithError(err error) *DomainError {
	e.Err = err
	return e
}

// Predefined domain errors
var (
	// Authentication errors
	AuthInvalidCredentials = NewDomainError("AUTH_INVALID_CREDENTIALS", "Invalid email or password", 401)
	AuthTokenExpired       = NewDomainError("AUTH_TOKEN_EXPIRED", "Authentication token has expired", 401)
	AuthTokenInvalid       = NewDomainError("AUTH_TOKEN_INVALID", "Authentication token is invalid", 401)
	AuthUnauthorized       = NewDomainError("AUTH_UNAUTHORIZED", "Authentication required", 401)
	AuthForbidden          = NewDomainError("AUTH_FORBIDDEN", "Access forbidden", 403)

	// User errors
	UserNotFound         = NewDomainError("USER_NOT_FOUND", "User not found", 404)
	UserAlreadyExists    = NewDomainError("USER_ALREADY_EXISTS", "User already exists", 409)
	UserInactive         = NewDomainError("USER_INACTIVE", "User account is inactive", 403)
	UserValidationFailed = NewDomainError("USER_VALIDATION_FAILED", "User validation failed", 400)

	// Video errors
	VideoNotFound         = NewDomainError("VIDEO_NOT_FOUND", "Video not found", 404)
	VideoNotReady         = NewDomainError("VIDEO_NOT_READY", "Video is not ready for streaming", 425)
	VideoProcessingFailed = NewDomainError("VIDEO_PROCESSING_FAILED", "Video processing failed", 500)
	VideoUploadFailed     = NewDomainError("VIDEO_UPLOAD_FAILED", "Video upload failed", 500)
	VideoAccessDenied     = NewDomainError("VIDEO_ACCESS_DENIED", "Access denied to video", 403)
	VideoValidationFailed = NewDomainError("VIDEO_VALIDATION_FAILED", "Video validation failed", 400)

	// Validation errors
	ValidationFailed     = NewDomainError("VALIDATION_FAILED", "Input validation failed", 400)
	InvalidFormat        = NewDomainError("INVALID_FORMAT", "Invalid format", 400)
	MissingRequiredField = NewDomainError("MISSING_REQUIRED_FIELD", "Required field is missing", 400)

	// System errors
	DatabaseError        = NewDomainError("DATABASE_ERROR", "Database operation failed", 500)
	CacheError           = NewDomainError("CACHE_ERROR", "Cache operation failed", 500)
	ExternalServiceError = NewDomainError("EXTERNAL_SERVICE_ERROR", "External service error", 502)
	InternalServerError  = NewDomainError("INTERNAL_SERVER_ERROR", "Internal server error", 500)

	// Rate limiting
	RateLimitExceeded = NewDomainError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded", 429)
)

// ValidationError represents field validation errors
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}

	if len(ve.Errors) == 1 {
		return fmt.Sprintf("validation failed: %s %s", ve.Errors[0].Field, ve.Errors[0].Message)
	}

	return fmt.Sprintf("validation failed: %d errors", len(ve.Errors))
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field, message string, value interface{}) {
	ve.Errors = append(ve.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// NewValidationErrors creates a new validation errors collection
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]ValidationError, 0),
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapWithCode wraps an error with a domain error
func WrapWithCode(err error, domainErr *DomainError) *DomainError {
	return domainErr.WithError(err)
}

// IsType checks if error is of specific domain error type
func IsType(err error, errorType string) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type == errorType
	}
	return false
}

// GetCode extracts error code from domain error
func GetCode(err error) int {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return 500 // Default to internal server error
}

// GetType extracts error type from domain error
func GetType(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type
	}
	return "UNKNOWN_ERROR"
}

// IsNotFound checks if error is a "not found" type error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrVideoNotFound) ||
		errors.Is(err, ErrRecordNotFound) ||
		IsType(err, "USER_NOT_FOUND") ||
		IsType(err, "VIDEO_NOT_FOUND")
}

// IsUnauthorized checks if error is an authorization error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrTokenExpired) ||
		errors.Is(err, ErrInvalidToken) ||
		IsType(err, "AUTH_UNAUTHORIZED") ||
		IsType(err, "AUTH_INVALID_CREDENTIALS") ||
		IsType(err, "AUTH_TOKEN_EXPIRED")
}

// IsForbidden checks if error is a forbidden access error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden) ||
		errors.Is(err, ErrVideoAccess) ||
		errors.Is(err, ErrInsufficientPermissions) ||
		IsType(err, "AUTH_FORBIDDEN") ||
		IsType(err, "VIDEO_ACCESS_DENIED")
}

// IsValidation checks if error is a validation error
func IsValidation(err error) bool {
	var validationErr *ValidationErrors
	return errors.As(err, &validationErr) ||
		errors.Is(err, ErrInvalidInput) ||
		IsType(err, "VALIDATION_FAILED")
}

// IsRateLimit checks if error is a rate limiting error
func IsRateLimit(err error) bool {
	return errors.Is(err, ErrRateLimitExceeded) ||
		errors.Is(err, ErrTooManyRequests) ||
		IsType(err, "RATE_LIMIT_EXCEEDED")
}
