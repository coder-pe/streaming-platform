// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package dbutil

import (
	"database/sql"
	"fmt"
)

// CheckRowsAffected verifica si se afectó al menos una fila
// Retorna el error especificado si no se afectaron filas
func CheckRowsAffected(result sql.Result, notFoundErr error) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return notFoundErr
	}

	return nil
}

// CheckExactlyOneRowAffected verifica que se haya afectado exactamente una fila
func CheckExactlyOneRowAffected(result sql.Result, notFoundErr error) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return notFoundErr
	}

	if rowsAffected > 1 {
		return fmt.Errorf("expected to affect 1 row, but affected %d rows", rowsAffected)
	}

	return nil
}

// ExecuteInTransaction ejecuta una función dentro de una transacción
// Si la función retorna error, hace rollback automáticamente
func ExecuteInTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ScanNullString convierte un sql.NullString a string
func ScanNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// ScanNullInt64 convierte un sql.NullInt64 a int64
func ScanNullInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

// ScanNullFloat64 convierte un sql.NullFloat64 a float64
func ScanNullFloat64(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

// ScanNullBool convierte un sql.NullBool a bool
func ScanNullBool(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}

// BuildPagination construye LIMIT y OFFSET para paginación
func BuildPagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	return limit, offset
}

// NormalizeSearchQuery normaliza una query de búsqueda para LIKE
func NormalizeSearchQuery(query string) string {
	return "%" + query + "%"
}
