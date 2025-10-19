// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package httputil

import (
	"net/http"
	"time"
)

const (
	// RefreshTokenCookieName es el nombre de la cookie para el refresh token
	RefreshTokenCookieName = "refresh_token"

	// RefreshTokenMaxAge es la duración máxima del refresh token (7 días)
	RefreshTokenMaxAge = 7 * 24 * 60 * 60 // 7 días en segundos
)

// CookieOptions representa las opciones para crear una cookie
type CookieOptions struct {
	Name     string
	Value    string
	Path     string
	MaxAge   int
	HttpOnly bool
	Secure   bool
	SameSite http.SameSite
}

// DefaultCookieOptions retorna las opciones por defecto para cookies seguras
func DefaultCookieOptions() CookieOptions {
	return CookieOptions{
		Path:     "/api/auth",
		MaxAge:   RefreshTokenMaxAge,
		HttpOnly: true,
		Secure:   true, // En producción debe ser true con HTTPS
		SameSite: http.SameSiteStrictMode,
	}
}

// SetCookie establece una cookie con las opciones especificadas
func SetCookie(w http.ResponseWriter, opts CookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     opts.Name,
		Value:    opts.Value,
		Path:     opts.Path,
		MaxAge:   opts.MaxAge,
		HttpOnly: opts.HttpOnly,
		Secure:   opts.Secure,
		SameSite: opts.SameSite,
	})
}

// SetRefreshTokenCookie establece la cookie del refresh token
func SetRefreshTokenCookie(w http.ResponseWriter, token string) {
	opts := DefaultCookieOptions()
	opts.Name = RefreshTokenCookieName
	opts.Value = token
	SetCookie(w, opts)
}

// ClearRefreshTokenCookie elimina la cookie del refresh token
func ClearRefreshTokenCookie(w http.ResponseWriter) {
	opts := DefaultCookieOptions()
	opts.Name = RefreshTokenCookieName
	opts.Value = ""
	opts.MaxAge = -1
	SetCookie(w, opts)
}

// GetRefreshTokenFromCookie obtiene el refresh token de la cookie
func GetRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshTokenCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// SetSessionCookie establece una cookie de sesión (sin MaxAge)
func SetSessionCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearCookie elimina una cookie específica
func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// SetSecureCookie establece una cookie con opciones de seguridad personalizadas
func SetSecureCookie(w http.ResponseWriter, name, value, path string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
	})
}
