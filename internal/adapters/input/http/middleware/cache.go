// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package middleware

import (
	"net/http"
	"strings"
)

// CacheControl configura headers de caché para archivos estáticos
func CacheControl() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// index.html y archivos HTML - No cachear (siempre obtener la versión más reciente)
			if strings.HasSuffix(path, ".html") || path == "/" {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
				next.ServeHTTP(w, r)
				return
			}

			// Archivos versionados (con ?v=) - Cachear por 1 año
			if r.URL.Query().Get("v") != "" {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				next.ServeHTTP(w, r)
				return
			}

			// CSS y JS sin versión - Cachear por 1 día (desarrollo)
			// En producción, estos siempre deberían tener versión
			if strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js") {
				w.Header().Set("Cache-Control", "public, max-age=86400") // 1 día
				next.ServeHTTP(w, r)
				return
			}

			// Imágenes, fuentes, videos - Cachear por 1 semana
			if isStaticAsset(path) {
				w.Header().Set("Cache-Control", "public, max-age=604800") // 1 semana
				next.ServeHTTP(w, r)
				return
			}

			// Por defecto - No cachear
			w.Header().Set("Cache-Control", "no-cache")
			next.ServeHTTP(w, r)
		})
	}
}

// isStaticAsset verifica si el path es un recurso estático
func isStaticAsset(path string) bool {
	staticExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp", // Imágenes
		".woff", ".woff2", ".ttf", ".eot", ".otf", // Fuentes
		".mp4", ".webm", ".ogg", ".mp3", // Media
		".ico", ".json", ".xml", // Otros
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
