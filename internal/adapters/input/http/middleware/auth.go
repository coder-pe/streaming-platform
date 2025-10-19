// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package middleware

import (
	"context"
	"net/http"
	"strings"

	"streaming-platform/pkg/jwt"
)

// JWTAuth middleware validates JWT tokens and sets user context
func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check if token has Bearer prefix
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := tokenParts[1]
			if tokenString == "" {
				http.Error(w, "Token required", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := jwt.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Set user context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "user_email", claims.Email)
			ctx = context.WithValue(ctx, "user_role", claims.Role)

			// Continue to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth middleware validates JWT tokens but doesn't require them
func OptionalAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No token provided, continue without auth
				next.ServeHTTP(w, r)
				return
			}

			// Check if token has Bearer prefix
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				// Invalid format, continue without auth
				next.ServeHTTP(w, r)
				return
			}

			tokenString := tokenParts[1]
			if tokenString == "" {
				// Empty token, continue without auth
				next.ServeHTTP(w, r)
				return
			}

			// Validate token
			claims, err := jwt.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				// Invalid token, continue without auth
				next.ServeHTTP(w, r)
				return
			}

			// Set user context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "user_email", claims.Email)
			ctx = context.WithValue(ctx, "user_role", claims.Role)

			// Continue to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("user_role")
			if userRole == nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			userRoleStr := userRole.(string)

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range roles {
				if userRoleStr == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware checks if user is admin
func RequireAdmin() func(http.Handler) http.Handler {
	return RequireRole("admin")
}

// RequireInstructor middleware checks if user is instructor or admin
func RequireInstructor() func(http.Handler) http.Handler {
	return RequireRole("instructor", "admin")
}

// AuthenticatedOnly middleware ensures user is authenticated
func AuthenticatedOnly() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("user_id")
			if userID == nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// ResourceOwner middleware checks if user owns the resource or is admin
func ResourceOwner(getOwnerID func(*http.Request) (string, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value("user_id")
			userRole := r.Context().Value("user_role")

			if userID == nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Admin can access any resource
			if userRole != nil && userRole.(string) == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// Get resource owner ID
			ownerID, err := getOwnerID(r)
			if err != nil {
				http.Error(w, "Invalid resource", http.StatusBadRequest)
				return
			}

			// Check if user owns the resource
			if userID.(string) != ownerID {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuth middleware for API key authentication
func APIKeyAuth(validAPIKeys []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// Try query parameter as fallback
				apiKey = r.URL.Query().Get("api_key")
			}

			if apiKey == "" {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}

			// Validate API key
			isValid := false
			for _, validKey := range validAPIKeys {
				if apiKey == validKey {
					isValid = true
					break
				}
			}

			if !isValid {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			// Set API key context for logging
			ctx := context.WithValue(r.Context(), "api_key", apiKey)

			// Continue to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SessionAuth middleware for session-based authentication
func SessionAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session cookie
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "Session required", http.StatusUnauthorized)
				return
			}

			if cookie.Value == "" {
				http.Error(w, "Invalid session", http.StatusUnauthorized)
				return
			}

			// TODO: Validate session with Redis or database
			// For now, just check if cookie exists

			// Set session context
			ctx := context.WithValue(r.Context(), "session_id", cookie.Value)

			// Continue to next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CSRFProtection middleware to prevent CSRF attacks
func CSRFProtection() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF check for safe methods
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Get CSRF token from header
			csrfToken := r.Header.Get("X-CSRF-Token")
			if csrfToken == "" {
				// Try form value as fallback
				csrfToken = r.FormValue("csrf_token")
			}

			if csrfToken == "" {
				http.Error(w, "CSRF token required", http.StatusForbidden)
				return
			}

			// TODO: Validate CSRF token
			// For now, just check if token exists

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; media-src 'self'")

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
