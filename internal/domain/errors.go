package domain

import "errors"

var (
	// ErrNotFound indicates the requested resource cannot be found.
	ErrNotFound = errors.New("resource not found")
	// ErrConflict indicates a unique constraint violation.
	ErrConflict = errors.New("resource conflict")
	// ErrInvalidInput represents a validation failure.
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnauthorized fired when auth fails.
	ErrUnauthorized = errors.New("unauthorized")
)
