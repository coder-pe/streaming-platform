// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package middleware

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SPAHandler sirve el index.html para todas las rutas que no son archivos estáticos o API
// Esto permite que las SPAs con routing del lado del cliente funcionen correctamente
func SPAHandler(staticPath string, indexPath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Si es una ruta de API, pasar al siguiente handler
			if strings.HasPrefix(path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}

			// Si es una ruta de archivos estáticos, pasar al siguiente handler
			if strings.HasPrefix(path, "/static/") {
				next.ServeHTTP(w, r)
				return
			}

			// Si es la raíz o una ruta SPA, verificar si el archivo existe
			fullPath := filepath.Join(staticPath, path)

			// Verificar si el archivo existe
			if stat, err := os.Stat(fullPath); err == nil {
				// Si es un directorio o el archivo existe, servir normalmente
				if !stat.IsDir() {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Para cualquier otra ruta (rutas SPA), servir el index.html
			http.ServeFile(w, r, indexPath)
		})
	}
}
