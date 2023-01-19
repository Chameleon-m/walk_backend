package repository

import (
	"errors"
)

var (
	// ErrInvalidString ...
	ErrInvalidString = errors.New("the provided string is not a valid ID")
	// ErrModelNotFound ...
	ErrModelNotFound = errors.New("model not found")
	// ErrModelUpdate ...
	ErrModelUpdate = errors.New("no upsert was done")
)

// IsErrInvalidString check is a ErrInvalidString
func IsErrInvalidString(err error) bool {
	return errors.Is(err, ErrInvalidString)
}

// IsErrModelNotFound check is a ErrModelNotFound
func IsErrModelNotFound(err error) bool {
	return errors.Is(err, ErrModelNotFound)
}

// IsErrModelUpdate check is a ErrModelUpdate
func IsErrModelUpdate(err error) bool {
	return errors.Is(err, ErrModelUpdate)
}
