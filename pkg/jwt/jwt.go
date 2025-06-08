package jwt

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID       string                 `json:"user_id"`
	Email        string                 `json:"email"`
	Role         string                 `json:"role"`
	CustomClaims map[string]interface{} `json:"custom_claims,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// GenerateToken creates a new JWT token with the given claims
func GenerateToken(claims *Claims, secret string, duration time.Duration) (string, error) {
	// Set expiration time
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(duration))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	claims.NotBefore = jwt.NewNumericDate(time.Now())
	claims.Issuer = "streaming-platform"

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString, secret string) (*Claims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token is expired")
	}

	// Check if token is not yet valid
	if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
		return nil, fmt.Errorf("token is not yet valid")
	}

	return claims, nil
}

// GenerateTokenPair creates both access and refresh tokens
func GenerateTokenPair(claims *Claims, secret string, accessDuration, refreshDuration time.Duration) (*TokenPair, error) {
	// Generate access token
	accessToken, err := GenerateToken(claims, secret, accessDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token with minimal claims
	refreshClaims := &Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		CustomClaims: map[string]interface{}{
			"type": "refresh",
		},
	}

	refreshToken, err := GenerateToken(refreshClaims, secret, refreshDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessDuration.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func RefreshAccessToken(refreshTokenString, secret string, newAccessDuration time.Duration) (string, error) {
	// Validate refresh token
	refreshClaims, err := ValidateToken(refreshTokenString, secret)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if it's actually a refresh token
	if tokenType, ok := refreshClaims.CustomClaims["type"].(string); !ok || tokenType != "refresh" {
		return "", fmt.Errorf("invalid token type for refresh")
	}

	// Create new access token claims
	newClaims := &Claims{
		UserID: refreshClaims.UserID,
		Email:  refreshClaims.Email,
		Role:   refreshClaims.Role,
	}

	// Generate new access token
	accessToken, err := GenerateToken(newClaims, secret, newAccessDuration)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return accessToken, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	// Check Bearer prefix
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("authorization header must start with 'Bearer '")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}

// GetTokenClaims extracts claims from a token without validation (for debugging)
func GetTokenClaims(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// IsTokenExpired checks if a token is expired without full validation
func IsTokenExpired(tokenString string) (bool, error) {
	claims, err := GetTokenClaims(tokenString)
	if err != nil {
		return true, err
	}

	if claims.ExpiresAt == nil {
		return false, nil
	}

	return claims.ExpiresAt.Time.Before(time.Now()), nil
}

// GetTokenTimeLeft returns the remaining time before token expiration
func GetTokenTimeLeft(tokenString string) (time.Duration, error) {
	claims, err := GetTokenClaims(tokenString)
	if err != nil {
		return 0, err
	}

	if claims.ExpiresAt == nil {
		return 0, fmt.Errorf("token has no expiration time")
	}

	timeLeft := time.Until(claims.ExpiresAt.Time)
	if timeLeft < 0 {
		return 0, nil // Token is expired
	}

	return timeLeft, nil
}

// CreateCustomToken creates a token with custom claims for specific purposes
func CreateCustomToken(userID, tokenType string, customData map[string]interface{}, secret string, duration time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		CustomClaims: map[string]interface{}{
			"type": tokenType,
		},
	}

	// Add custom data to claims
	for key, value := range customData {
		claims.CustomClaims[key] = value
	}

	return GenerateToken(claims, secret, duration)
}

// CreatePasswordResetToken creates a token for password reset
func CreatePasswordResetToken(userID, email, secret string) (string, error) {
	customData := map[string]interface{}{
		"email": email,
	}

	return CreateCustomToken(userID, "password_reset", customData, secret, 1*time.Hour)
}

// CreateEmailVerificationToken creates a token for email verification
func CreateEmailVerificationToken(userID, email, secret string) (string, error) {
	customData := map[string]interface{}{
		"email": email,
	}

	return CreateCustomToken(userID, "email_verification", customData, secret, 24*time.Hour)
}

// CreateStreamingToken creates a token for video streaming access
func CreateStreamingToken(userID, videoID, secret string, duration time.Duration) (string, error) {
	customData := map[string]interface{}{
		"video_id": videoID,
	}

	return CreateCustomToken(userID, "streaming", customData, secret, duration)
}

// ValidatePasswordResetToken validates a password reset token
func ValidatePasswordResetToken(tokenString, secret string) (userID, email string, err error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return "", "", err
	}

	// Check token type
	tokenType, ok := claims.CustomClaims["type"].(string)
	if !ok || tokenType != "password_reset" {
		return "", "", fmt.Errorf("invalid token type for password reset")
	}

	// Extract email
	emailClaim, ok := claims.CustomClaims["email"].(string)
	if !ok {
		return "", "", fmt.Errorf("email not found in token")
	}

	return claims.UserID, emailClaim, nil
}

// ValidateEmailVerificationToken validates an email verification token
func ValidateEmailVerificationToken(tokenString, secret string) (userID, email string, err error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return "", "", err
	}

	// Check token type
	tokenType, ok := claims.CustomClaims["type"].(string)
	if !ok || tokenType != "email_verification" {
		return "", "", fmt.Errorf("invalid token type for email verification")
	}

	// Extract email
	emailClaim, ok := claims.CustomClaims["email"].(string)
	if !ok {
		return "", "", fmt.Errorf("email not found in token")
	}

	return claims.UserID, emailClaim, nil
}

// ValidateStreamingToken validates a streaming access token
func ValidateStreamingToken(tokenString, secret string) (userID, videoID string, err error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return "", "", err
	}

	// Check token type
	tokenType, ok := claims.CustomClaims["type"].(string)
	if !ok || tokenType != "streaming" {
		return "", "", fmt.Errorf("invalid token type for streaming")
	}

	// Extract video ID
	videoIDClaim, ok := claims.CustomClaims["video_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("video ID not found in token")
	}

	return claims.UserID, videoIDClaim, nil
}

// RevokeToken adds a token to a blacklist (requires external storage like Redis)
type TokenBlacklist interface {
	Add(tokenID string, expiration time.Time) error
	IsBlacklisted(tokenID string) (bool, error)
}

// BlacklistToken adds a token to the blacklist
func BlacklistToken(tokenString string, blacklist TokenBlacklist) error {
	claims, err := GetTokenClaims(tokenString)
	if err != nil {
		return fmt.Errorf("failed to extract claims for blacklisting: %w", err)
	}

	// Use JWT ID if available, otherwise use a hash of the token
	tokenID := claims.ID
	if tokenID == "" {
		tokenID = fmt.Sprintf("%x", sha256.Sum256([]byte(tokenString)))
	}

	expiration := time.Now().Add(24 * time.Hour) // Default blacklist duration
	if claims.ExpiresAt != nil {
		expiration = claims.ExpiresAt.Time
	}

	return blacklist.Add(tokenID, expiration)
}

// IsTokenBlacklisted checks if a token is blacklisted
func IsTokenBlacklisted(tokenString string, blacklist TokenBlacklist) (bool, error) {
	claims, err := GetTokenClaims(tokenString)
	if err != nil {
		return false, fmt.Errorf("failed to extract claims for blacklist check: %w", err)
	}

	tokenID := claims.ID
	if tokenID == "" {
		tokenID = fmt.Sprintf("%x", sha256.Sum256([]byte(tokenString)))
	}

	return blacklist.IsBlacklisted(tokenID)
}
