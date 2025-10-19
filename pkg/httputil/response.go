// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package httputil

import (
	"encoding/json"
	"net/http"
)

// Response representa una respuesta JSON estándar
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// WriteJSON escribe una respuesta JSON con el código de estado especificado
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// WriteSuccess escribe una respuesta exitosa con datos
func WriteSuccess(w http.ResponseWriter, status int, data interface{}) error {
	return WriteJSON(w, status, Response{
		Success: true,
		Data:    data,
	})
}

// WriteSuccessMessage escribe una respuesta exitosa con solo un mensaje
func WriteSuccessMessage(w http.ResponseWriter, status int, message string) error {
	return WriteJSON(w, status, Response{
		Success: true,
		Message: message,
	})
}

// WriteError escribe una respuesta de error
func WriteError(w http.ResponseWriter, status int, message string) error {
	return WriteJSON(w, status, Response{
		Success: false,
		Error:   message,
	})
}

// WriteValidationError escribe un error de validación (400)
func WriteValidationError(w http.ResponseWriter, message string) error {
	return WriteError(w, http.StatusBadRequest, message)
}

// WriteUnauthorized escribe un error de autenticación (401)
func WriteUnauthorized(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Unauthorized"
	}
	return WriteError(w, http.StatusUnauthorized, message)
}

// WriteForbidden escribe un error de acceso prohibido (403)
func WriteForbidden(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Forbidden"
	}
	return WriteError(w, http.StatusForbidden, message)
}

// WriteNotFound escribe un error de recurso no encontrado (404)
func WriteNotFound(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Resource not found"
	}
	return WriteError(w, http.StatusNotFound, message)
}

// WriteInternalError escribe un error interno del servidor (500)
func WriteInternalError(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Internal server error"
	}
	return WriteError(w, http.StatusInternalServerError, message)
}

// DecodeJSON decodifica el body de una petición JSON
func DecodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return json.NewDecoder(http.NoBody).Decode(v)
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
