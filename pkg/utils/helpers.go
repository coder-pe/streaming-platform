package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// StringUtils provides string utility functions
type StringUtils struct{}

// IsEmpty checks if a string is empty or contains only whitespace
func (s StringUtils) IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// Truncate truncates a string to a specified length
func (s StringUtils) Truncate(str string, length int) string {
	if len(str) <= length {
		return str
	}

	if length <= 3 {
		return str[:length]
	}

	return str[:length-3] + "..."
}

// ToSnakeCase converts a string to snake_case
func (s StringUtils) ToSnakeCase(str string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

// ToCamelCase converts a string to camelCase
func (s StringUtils) ToCamelCase(str string) string {
	words := strings.FieldsFunc(str, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	if len(words) == 0 {
		return ""
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			result += strings.ToUpper(string(words[i][0])) + strings.ToLower(words[i][1:])
		}
	}

	return result
}

// Slugify creates a URL-friendly slug from a string
func (s StringUtils) Slugify(str string) string {
	// Convert to lowercase
	str = strings.ToLower(str)

	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile("[^a-z0-9]+")
	str = re.ReplaceAllString(str, "-")

	// Remove leading and trailing hyphens
	str = strings.Trim(str, "-")

	return str
}

// Contains checks if a slice contains a specific string
func (s StringUtils) Contains(slice []string, item string) bool {
	for _, str := range slice {
		if str == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func (s StringUtils) RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, str := range slice {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}

// RandomString generates a random string of specified length
func (s StringUtils) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[randomByte[0]%byte(len(charset))]
	}
	return string(b)
}

// ValidationUtils provides validation utility functions
type ValidationUtils struct{}

// IsValidEmail validates an email address
func (v ValidationUtils) IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidURL validates a URL
func (v ValidationUtils) IsValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// IsValidUUID validates a UUID string
func (v ValidationUtils) IsValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

// IsStrongPassword checks if a password meets strength requirements
func (v ValidationUtils) IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// IsValidPhoneNumber validates a phone number (basic validation)
func (v ValidationUtils) IsValidPhoneNumber(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(strings.ReplaceAll(phone, " ", ""))
}

// FileUtils provides file utility functions
type FileUtils struct{}

// GetFileExtension returns the file extension
func (f FileUtils) GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// IsImageFile checks if a file is an image based on extension
func (f FileUtils) IsImageFile(filename string) bool {
	ext := f.GetFileExtension(filename)
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"}

	for _, imageExt := range imageExts {
		if ext == imageExt {
			return true
		}
	}
	return false
}

// IsVideoFile checks if a file is a video based on extension
func (f FileUtils) IsVideoFile(filename string) bool {
	ext := f.GetFileExtension(filename)
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm", ".m4v"}

	for _, videoExt := range videoExts {
		if ext == videoExt {
			return true
		}
	}
	return false
}

// FormatFileSize formats file size in human-readable format
func (f FileUtils) FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GenerateUniqueFilename generates a unique filename with timestamp and random string
func (f FileUtils) GenerateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Format("20060102_150405")

	stringUtils := StringUtils{}
	randomStr := stringUtils.RandomString(8)

	return fmt.Sprintf("%s_%s_%s%s", name, timestamp, randomStr, ext)
}

// TimeUtils provides time utility functions
type TimeUtils struct{}

// FormatDuration formats duration in human-readable format
func (t TimeUtils) FormatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}

	minutes := seconds / 60
	remainingSeconds := seconds % 60

	if minutes < 60 {
		if remainingSeconds == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm %ds", minutes, remainingSeconds)
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60

	if remainingMinutes == 0 && remainingSeconds == 0 {
		return fmt.Sprintf("%dh", hours)
	} else if remainingSeconds == 0 {
		return fmt.Sprintf("%dh %dm", hours, remainingMinutes)
	}

	return fmt.Sprintf("%dh %dm %ds", hours, remainingMinutes, remainingSeconds)
}

// TimeAgo returns a human-readable time difference
func (t TimeUtils) TimeAgo(timestamp time.Time) string {
	now := time.Now()
	diff := now.Sub(timestamp)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, pluralize(minutes))
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, pluralize(hours))
	} else if diff < 30*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d day%s ago", days, pluralize(days))
	} else if diff < 12*30*24*time.Hour {
		months := int(diff.Hours() / (24 * 30))
		return fmt.Sprintf("%d month%s ago", months, pluralize(months))
	} else {
		years := int(diff.Hours() / (24 * 365))
		return fmt.Sprintf("%d year%s ago", years, pluralize(years))
	}
}

// IsBusinessDay checks if a given date is a business day (Monday-Friday)
func (t TimeUtils) IsBusinessDay(date time.Time) bool {
	weekday := date.Weekday()
	return weekday >= time.Monday && weekday <= time.Friday
}

// GetStartOfDay returns the start of the day for a given time
func (t TimeUtils) GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay returns the end of the day for a given time
func (t TimeUtils) GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// CryptoUtils provides cryptographic utility functions
type CryptoUtils struct{}

// GenerateRandomBytes generates random bytes
func (c CryptoUtils) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

// GenerateRandomHex generates a random hex string
func (c CryptoUtils) GenerateRandomHex(length int) (string, error) {
	bytes, err := c.GenerateRandomBytes(length / 2)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashSHA256 creates a SHA256 hash of the input
func (c CryptoUtils) HashSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// GenerateSecureToken generates a secure random token
func (c CryptoUtils) GenerateSecureToken(length int) (string, error) {
	return c.GenerateRandomHex(length)
}

// HTTPUtils provides HTTP utility functions
type HTTPUtils struct{}

// GetClientIP extracts the real client IP from request
func (h HTTPUtils) GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in case of multiple proxies
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// GetUserAgent extracts user agent from request
func (h HTTPUtils) GetUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// IsAjaxRequest checks if the request is an AJAX request
func (h HTTPUtils) IsAjaxRequest(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// GetBearerToken extracts bearer token from Authorization header
func (h HTTPUtils) GetBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return ""
	}

	return authHeader[len(bearerPrefix):]
}

// SetJSONResponse sets JSON content type header
func (h HTTPUtils) SetJSONResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// SetCORSHeaders sets CORS headers
func (h HTTPUtils) SetCORSHeaders(w http.ResponseWriter, origin string) {
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// MathUtils provides mathematical utility functions
type MathUtils struct{}

// RoundToDecimalPlace rounds a float to specified decimal places
func (m MathUtils) RoundToDecimalPlace(value float64, places int) float64 {
	multiplier := math.Pow(10, float64(places))
	return math.Round(value*multiplier) / multiplier
}

// Percentage calculates percentage
func (m MathUtils) Percentage(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (part / total) * 100
}

// Clamp clamps a value between min and max
func (m MathUtils) Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ArrayUtils provides array utility functions
type ArrayUtils struct{}

// ChunkStrings splits a string slice into chunks of specified size
func (a ArrayUtils) ChunkStrings(slice []string, chunkSize int) [][]string {
	if chunkSize <= 0 {
		return nil
	}

	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// FilterStrings filters a string slice based on a predicate function
func (a ArrayUtils) FilterStrings(slice []string, predicate func(string) bool) []string {
	var result []string
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// MapStrings transforms a string slice using a mapper function
func (a ArrayUtils) MapStrings(slice []string, mapper func(string) string) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

// ConversionUtils provides type conversion utilities
type ConversionUtils struct{}

// StringToInt converts string to int with default value
func (c ConversionUtils) StringToInt(s string, defaultValue int) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultValue
}

// StringToInt64 converts string to int64 with default value
func (c ConversionUtils) StringToInt64(s string, defaultValue int64) int64 {
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val
	}
	return defaultValue
}

// StringToFloat64 converts string to float64 with default value
func (c ConversionUtils) StringToFloat64(s string, defaultValue float64) float64 {
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return defaultValue
}

// StringToBool converts string to bool with default value
func (c ConversionUtils) StringToBool(s string, defaultValue bool) bool {
	if val, err := strconv.ParseBool(s); err == nil {
		return val
	}
	return defaultValue
}

// Helper functions

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// Global utility instances for easy access
var (
	String     = StringUtils{}
	Validation = ValidationUtils{}
	File       = FileUtils{}
	Time       = TimeUtils{}
	Crypto     = CryptoUtils{}
	HTTP       = HTTPUtils{}
	Math       = MathUtils{}
	Array      = ArrayUtils{}
	Conversion = ConversionUtils{}
)

// Pagination utilities
type PaginationResult struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, total int64) PaginationResult {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return PaginationResult{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// GetOffset calculates the offset for pagination
func GetOffset(page, limit int) int {
	if page <= 0 {
		page = 1
	}
	return (page - 1) * limit
}

// Response utilities
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SuccessResponse creates a success API response
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// ErrorResponse creates an error API response
func ErrorResponse(message string) APIResponse {
	return APIResponse{
		Success: false,
		Error:   message,
	}
}

// MessageResponse creates a message API response
func MessageResponse(message string) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
	}
}

// Environment utilities
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
