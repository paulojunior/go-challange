package services

import "errors"

// ErrNotFound indicates that the requested resource was not found.
var ErrNotFound = errors.New("resource not found")

// ErrInvalidInput indicates that the provided input is invalid.
var ErrInvalidInput = errors.New("invalid input")

// Specific validation errors
var (
	ErrInvalidOffset        = errors.New("offset must be a non-negative integer")
	ErrInvalidLimit         = errors.New("limit must be a positive integer")
	ErrInvalidPrice         = errors.New("priceLessThan must be a valid decimal number")
	ErrNegativePrice        = errors.New("priceLessThan must be a non-negative value")
	ErrInvalidCategoryInput = errors.New("category code and name are required")
)
