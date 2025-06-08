package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	AllowCredentials   bool
	MaxAge             int
	OptionsPassthrough bool
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Accept-Language",
			"Content-Language",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Content-Range",
			"Accept-Ranges",
		},
		AllowCredentials:   true,
		MaxAge:             86400, // 24 hours
		OptionsPassthrough: false,
	}
}

// ProductionCORSConfig returns a production-ready CORS configuration
func ProductionCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{
			"https://streamlearn.com",
			"https://www.streamlearn.com",
			"https://app.streamlearn.com",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Accept-Language",
			"Content-Language",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials:   true,
		MaxAge:             3600, // 1 hour
		OptionsPassthrough: false,
	}
}

// CORS middleware with default configuration
func CORS() func(http.Handler) http.Handler {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig middleware with custom configuration
func CORSWithConfig(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if origin != "" && isOriginAllowed(origin, config.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// Set credentials header
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				// Set allowed methods
				if len(config.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				}

				// Set allowed headers
				if len(config.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				}

				// Set max age
				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				}

				// If OptionsPassthrough is false, handle preflight here
				if !config.OptionsPassthrough {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			// Set exposed headers for actual requests
			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if origin is in allowed origins list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Check for wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}
	return false
}

// RestrictiveCORS middleware for sensitive endpoints
func RestrictiveCORS(allowedOrigins []string) func(http.Handler) http.Handler {
	config := CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           3600,
	}
	return CORSWithConfig(config)
}

// APICorsMiddleware specifically for API endpoints
func APICorsMiddleware() func(http.Handler) http.Handler {
	config := CORSConfig{
		AllowedOrigins: []string{"*"}, // In production, specify exact origins
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
			"X-Requested-With",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Total-Count",
			"X-Page",
			"X-Per-Page",
		},
		AllowCredentials: false, // API typically doesn't need credentials
		MaxAge:           86400,
	}
	return CORSWithConfig(config)
}

// StreamingCORS middleware for video streaming endpoints
func StreamingCORS() func(http.Handler) http.Handler {
	config := CORSConfig{
		AllowedOrigins: []string{"*"}, // Streaming content often needs broad access
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Range",
			"Accept",
			"Accept-Encoding",
			"Authorization",
			"Content-Type",
		},
		ExposedHeaders: []string{
			"Accept-Ranges",
			"Content-Range",
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           86400,
	}
	return CORSWithConfig(config)
}

// WebSocketCORS middleware for WebSocket connections
func WebSocketCORS(allowedOrigins []string) func(http.Handler) http.Handler {
	config := CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
		},
		AllowedHeaders: []string{
			"Origin",
			"Sec-WebSocket-Protocol",
			"Sec-WebSocket-Extensions",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           3600,
	}
	return CORSWithConfig(config)
}

// NoCache middleware adds cache prevention headers
func NoCache() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			next.ServeHTTP(w, r)
		})
	}
}

// CacheHeaders middleware adds appropriate cache headers
func CacheHeaders(maxAge int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxAge > 0 {
				w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
			}
			next.ServeHTTP(w, r)
		})
	}
}
